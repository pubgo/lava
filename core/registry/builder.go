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
	"github.com/pubgo/lava/core/runmode"
	"github.com/pubgo/lava/internal/pkg/netutil"
	"github.com/pubgo/lava/internal/pkg/syncx"
	"github.com/pubgo/lava/logging/logutil"
	"github.com/pubgo/lava/version"
)

func init() {
	dix.Register(func(c config.Config) *Cfg {
		var cfg = DefaultCfg()

		// 配置解析
		xerror.Panic(c.UnmarshalKey(Name, &cfg))
		return cfg.Check()
	})

	dix.Register(func(lifecycle lifecycle.Lifecycle, app *config.App, cfg *Cfg, regs map[string]Registry) *Loader {
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

			register(reg, app)

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
						register(reg, app)
					case <-ctx.Done():
						zap.L().Info("service register cancelled")
						return
					}
				}
			})

			// 服务撤销
			lifecycle.BeforeStops(func() {
				cancel()
				deregister(reg, app)
			})
		})
		return new(Loader)
	})
}

func register(reg Registry, app *config.App) {
	// parse address for host, port
	var advt, host string
	var port = addPort(app.Addr)
	advt = app.Advertise

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
		Id:       runmode.Project + "-" + runmode.Hostname + "-" + runmode.InstanceID,
		Metadata: map[string]string{"registry": reg.String()},
	}

	s := &Service{
		Name:  runmode.Project,
		Nodes: []*Node{node},
	}

	logutil.OkOrErr(
		zap.L(),
		"register service node",
		func() error { return reg.Register(s) },
		zap.String("instance_id", node.Id),
		zap.String("service", runmode.Project),
		zap.String("registry", Default().String()),
	)
}

func deregister(reg Registry, app *config.App) {
	var advt, host string
	var port = addPort(app.Addr)
	advt = app.Advertise

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
		Id:       runmode.Project + "-" + runmode.Hostname + "-" + runmode.InstanceID,
		Metadata: make(map[string]string),
	}

	s := &Service{
		Name:  runmode.Project,
		Nodes: []*Node{node},
	}

	logutil.OkOrErr(
		zap.L(),
		"deregister service node",
		func() error { return reg.Deregister(s) },
		zap.String("id", node.Id),
		zap.String("name", runmode.Project),
	)
}
