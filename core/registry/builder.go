package registry

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/pubgo/dix"
	"github.com/pubgo/xerror"
	"go.uber.org/zap"

	"github.com/pubgo/lava/config"
	"github.com/pubgo/lava/core/lifecycle"
	"github.com/pubgo/lava/logging/logutil"
	"github.com/pubgo/lava/pkg/netutil"
	"github.com/pubgo/lava/pkg/syncx"
	"github.com/pubgo/lava/runtime"
	"github.com/pubgo/lava/service"
	"github.com/pubgo/lava/version"
)

func init() {
	dix.Register(func(c config.Config) *Cfg {
		var cfg = DefaultCfg()

		// 配置解析
		xerror.Panic(c.UnmarshalKey(Name, &cfg))
		return cfg.Check()
	})

	dix.Register(func(lifecycle lifecycle.Lifecycle, app service.AppInfo, cfg *Cfg, regs map[string]Registry) *Loader {
		reg := regs[cfg.Driver]
		xerror.AssertFn(reg == nil, func() error {
			var errs = fmt.Errorf("registry driver is null")
			errs = xerror.WrapF(errs, "driver=>%s", cfg.Driver)
			errs = xerror.WrapF(errs, "regs=>%v", regs)
			return errs
		})

		// 服务注册
		lifecycle.AfterStarts(func() {
			SetDefault(reg)

			xerror.Panic(Register(reg, app))

			var cancel = syncx.GoCtx(func(ctx context.Context) {
				var interval = DefaultRegisterInterval

				if cfg.RegisterInterval > time.Duration(0) {
					interval = cfg.RegisterInterval
				}

				var tick = time.NewTicker(interval)
				defer tick.Stop()

				for {
					select {
					case <-tick.C:
						logutil.LogOrErr(zap.L(), "service Register",
							func() error { return Register(reg, app) },
							zap.String("service", app.Options().Name),
							zap.String("instanceId", app.Options().Id),
							zap.String("registry", Default().String()),
							zap.String("interval", interval.String()),
						)
					case <-ctx.Done():
						zap.L().Info("service Register cancelled")
						return
					}
				}
			})

			// 服务撤销
			lifecycle.BeforeStops(func() {
				cancel()
				xerror.Panic(Deregister(reg, app))
			})
		})
		return new(Loader)
	})
}

func Register(reg Registry, app service.AppInfo) (err error) {
	defer xerror.RecoverErr(&err, func(err xerror.XErr) xerror.XErr {
		return err.WrapF("Register service=>%#v", app.Options())
	})

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

	// Register service
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
		"Register service node",
		func() error { return reg.Register(s) },
		zap.String("id", node.Id),
		zap.String("name", opt.Name))
	return nil
}

func Deregister(reg Registry, app service.AppInfo) (err error) {
	defer xerror.RecoverErr(&err, func(err xerror.XErr) xerror.XErr {
		return err.WrapF("Deregister service=>%#v", app.Options())
	})

	var opt = app.Options()

	var advt, host string
	var port = opt.Port

	if len(opt.Advertise) > 0 {
		advt = opt.Advertise
	} else {
		advt = opt.Addr
	}

	parts := strings.Split(advt, ":")
	if len(parts) > 1 {
		host = strings.Join(parts[:len(parts)-1], ":")
		port, _ = strconv.Atoi(parts[len(parts)-1])
	} else {
		host = parts[0]
	}

	// Register service
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

	logutil.LogOrErr(zap.L(), "Deregister service node",
		func() error { return reg.Deregister(s) },
		zap.String("id", node.Id),
		zap.String("name", opt.Name),
	)

	return nil
}
