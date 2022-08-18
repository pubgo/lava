package registry

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/pubgo/xerror"
	"go.uber.org/zap"

	"github.com/pubgo/lava/core/lifecycle"
	"github.com/pubgo/lava/core/runmode"
	"github.com/pubgo/lava/internal/pkg/netutil"
	"github.com/pubgo/lava/internal/pkg/syncx"
	"github.com/pubgo/lava/logging/logutil"
	"github.com/pubgo/lava/version"
)

func New(c *Cfg, lifecycle lifecycle.Lifecycle, regs map[string]Registry) {
	var cfg = DefaultCfg()

	// 配置解析
	cfg.Check()

	reg := regs[cfg.Driver]
	xerror.AssertFn(reg == nil, func() error {
		var errs = fmt.Errorf("registry driver is null")
		errs = xerror.WrapF(errs, "driver=>%s", cfg.Driver)
		errs = xerror.WrapF(errs, "regs=>%v", regs)
		return errs
	})

	// 服务注册
	lifecycle.AfterStart(func() error {
		SetDefault(reg)

		register(reg)

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
					register(reg)
				case <-ctx.Done():
					zap.L().Info("service register cancelled")
					return
				}
			}
		})

		// 服务撤销
		lifecycle.BeforeStop(func() error {
			cancel()
			deregister(reg)
			return nil
		})
		return nil
	})
}

func register(reg Registry) {
	// parse address for host, port
	var advt, host string
	var port = runmode.GrpcPort

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
		Version:  version.Version(),
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

func deregister(reg Registry) {
	var advt, host string
	var port = runmode.GrpcPort

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
