package registry_plugin

import (
	"context"
	"fmt"
	"github.com/pubgo/lava/config"
	"github.com/pubgo/lava/core/logging/logutil"
	registry2 "github.com/pubgo/lava/core/registry"
	registry_type2 "github.com/pubgo/lava/core/registry/registry_type"
	"strconv"
	"strings"
	"time"

	"github.com/pubgo/xerror"
	"go.uber.org/zap"

	"github.com/pubgo/lava/pkg/netutil"
	"github.com/pubgo/lava/pkg/syncx"
	"github.com/pubgo/lava/plugin"
	"github.com/pubgo/lava/runtime"
	"github.com/pubgo/lava/service/service_type"
	"github.com/pubgo/lava/version"
)

const (
	// DefaultMaxMsgSize define maximum message size that server can send or receive.
	// Default value is 4MB.
	DefaultMaxMsgSize = 1024 * 1024 * 4

	DefaultSleepAfterDeRegister = time.Second * 2

	// DefaultRegisterTTL The register expiry time
	DefaultRegisterTTL = time.Minute

	// DefaultRegisterInterval The interval on which to register
	DefaultRegisterInterval = time.Second * 30

	defaultContentType = "application/grpc"

	DefaultSleepAfterDeregister = time.Second * 2
)

func Enable(srv service_type.Service) {
	srv.Plugin(&plugin.Base{
		Name: registry2.Name,
		OnInit: func(p plugin.Process) {
			var cfg = registry2.DefaultCfg()

			// 配置解析
			p.BeforeStart(func() {
				xerror.Panic(config.GetMap(registry2.Name).Decode(&cfg))
			})

			// 服务注册
			p.AfterStart(func() {
				registry2.DefaultRegistry = xerror.PanicErr(cfg.Build()).(registry_type2.Registry)
				registry2.DefaultRegistry.Init()

				xerror.Panic(register(srv))

				var cancel = syncx.GoCtx(func(ctx context.Context) {
					var interval = DefaultRegisterInterval

					// only process if it exists
					if cfg.RegisterInterval > time.Duration(0) {
						interval = cfg.RegisterInterval
					}

					var tick = time.NewTicker(interval)
					defer tick.Stop()

					for {
						select {
						case <-tick.C:
							logutil.LogOrErr(zap.L(), "service register",
								func() error { return register(srv) },
								zap.String("registry", registry2.Default().String()),
								zap.String("interval", interval.String()),
							)
						case <-ctx.Done():
							zap.L().Info("service register cancelled")
							return
						}
					}
				})

				// 服务撤销
				p.BeforeStop(func() {
					cancel()
					xerror.Panic(deregister(srv))
				})
			})
		},
	})
}

func register(srv service_type.Service) (err error) {
	defer xerror.RespErr(&err)

	var reg = registry2.Default()
	var opt = srv.Options()

	// parse address for host, port
	var advt, host string
	var port = opt.Port

	if len(opt.Advertise) > 0 {
		advt = opt.Advertise
	} else {
		advt = opt.Advertise
	}

	parts := strings.Split(advt, ":")
	if len(parts) > 1 {
		host = strings.Join(parts[:len(parts)-1], ":")
		port, _ = strconv.Atoi(parts[len(parts)-1])
	} else {
		host = parts[0]
	}

	if host == "" {
		host = netutil.GetLocalIP()
	}

	// register service
	node := &registry_type2.Node{
		Port:     port,
		Address:  fmt.Sprintf("%s:%d", host, port),
		Id:       opt.Name + "-" + runtime.Hostname + "-" + opt.Id,
		Metadata: make(map[string]string),
	}

	node.Metadata["registry"] = reg.String()

	services := &registry_type2.Service{
		Name:    opt.Name,
		Version: version.Version,
		Nodes:   []*registry_type2.Node{node},
	}

	zap.L().Info("Registering Node", zap.String("id", node.Id), zap.String("name", opt.Name))

	logutil.LogOrPanic(zap.L(), "[grpc] register", func() error { return reg.Register(services) })
	return nil
}

func deregister(srv service_type.Service) (err error) {
	defer xerror.RespErr(&err)

	var opt = srv.Options()
	var reg = registry2.Default()

	var advt, host string
	var port = opt.Port

	if len(opt.Advertise) > 0 {
		advt = opt.Advertise
	} else {
		advt = opt.Address
	}

	parts := strings.Split(advt, ":")
	if len(parts) > 1 {
		host = strings.Join(parts[:len(parts)-1], ":")
		port, _ = strconv.Atoi(parts[len(parts)-1])
	} else {
		host = parts[0]
	}

	// register service
	node := &registry_type2.Node{
		Port:     port,
		Address:  fmt.Sprintf("%s:%d", host, port),
		Id:       opt.Name + "-" + runtime.Hostname + "-" + opt.Id,
		Metadata: make(map[string]string),
	}

	services := &registry_type2.Service{
		Name:    opt.Name,
		Version: version.Version,
		Nodes:   []*registry_type2.Node{node},
	}

	logutil.LogOrErr(zap.L(), "deregister node",
		func() error { return reg.Deregister(services) },
		zap.String("id", node.Id),
	)

	return nil
}
