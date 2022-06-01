package registry

import (
	"context"
	"fmt"
	"github.com/pubgo/dix"
	"strconv"
	"strings"
	"time"

	"github.com/pubgo/xerror"
	"go.uber.org/zap"

	"github.com/pubgo/lava/config"
	"github.com/pubgo/lava/logging/logutil"
	"github.com/pubgo/lava/pkg/netutil"
	"github.com/pubgo/lava/pkg/syncx"
	"github.com/pubgo/lava/runtime"
	"github.com/pubgo/lava/service"
	"github.com/pubgo/lava/version"
)

func init() {
	dix.Register(func() *Cfg {
		var cfg = DefaultCfg()
		// 配置解析
		xerror.Panic(config.UnmarshalKey(Name, &cfg))
		return cfg.Check()
	})

	dix.Register(func(app service.App, cfg *Cfg, regs map[string]Registry) {
		reg := regs[cfg.Driver]
		var errs = xerror.AssertErr(reg == nil, "registry driver is null")
		errs = xerror.WrapF(errs, "driver=>%s", cfg.Driver)
		errs = xerror.WrapF(errs, "regs=>%v", regs)
		xerror.Panic(errs)

		// 服务注册
		app.AfterStarts(func() {
			reg.Init()

			SetDefault(reg)

			xerror.Panic(register(app))

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
							func() error { return register(app) },
							zap.String("service", app.Options().Name),
							zap.String("InstanceId", app.Options().Id),
							zap.String("registry", Default().String()),
							zap.String("interval", interval.String()),
						)
					case <-ctx.Done():
						zap.L().Info("service register cancelled")
						return
					}
				}
			})

			// 服务撤销
			app.BeforeStops(func() {
				cancel()
				xerror.Panic(deregister(app))
			})
		})
	})
}

func register(app service.App) (err error) {
	defer xerror.RespErr(&err)

	var reg = Default()
	var opt = app.Options()

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
	node := &Node{
		Port:     port,
		Version:  version.Version,
		Address:  fmt.Sprintf("%s:%d", host, port),
		Id:       opt.Name + "-" + runtime.Hostname + "-" + opt.Id,
		Metadata: map[string]string{"registry": reg.String()},
	}

	s := &Service{
		Name:  opt.Name,
		Nodes: []*Node{node},
	}

	logutil.LogOrPanic(
		zap.L(),
		"register service node",
		func() error { return reg.Register(s) },
		zap.String("id", node.Id),
		zap.String("name", opt.Name))
	return nil
}

func deregister(app service.App) (err error) {
	defer xerror.RespErr(&err)

	var opt = app.Options()
	var reg = Default()

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
	node := &Node{
		Port:     port,
		Address:  fmt.Sprintf("%s:%d", host, port),
		Id:       opt.Name + "-" + runtime.Hostname + "-" + opt.Id,
		Metadata: make(map[string]string),
	}

	s := &Service{
		Name:  opt.Name,
		Nodes: []*Node{node},
	}

	logutil.LogOrErr(zap.L(), "deregister service node",
		func() error { return reg.Deregister(s) },
		zap.String("id", node.Id),
		zap.String("name", opt.Name),
	)

	return nil
}
