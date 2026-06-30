package registry

import (
	"context"
	"fmt"
	"time"

	"github.com/pubgo/funk/v2/assert"
	"github.com/pubgo/funk/v2/async"
	"github.com/pubgo/funk/v2/buildinfo/version"
	"github.com/pubgo/funk/v2/errors"
	"github.com/pubgo/funk/v2/log"

	"github.com/pubgo/lava/v2/core/lifecycle"
	"github.com/pubgo/lava/v2/core/running"
	"github.com/pubgo/lava/v2/core/service"
	"github.com/pubgo/lava/v2/internal/logutil"
	"github.com/pubgo/lava/v2/pkg/netutil"
)

// New 注册 lifecycle 钩子：启动后向注册中心注册服务并定期续约，停止前撤销注册。
func New(c *Config, lifecycle lifecycle.Lifecycle, regs map[string]Registry) {
	cfg := DefaultCfg()

	cfg.Check()

	reg := regs[cfg.Driver]
	assert.Fn(reg == nil, func() error {
		return &errors.Err{
			Msg: "registry driver is null",
			Tags: errors.Tags{
				"driver": cfg.Driver,
				"regs":   regs,
			},
		}
	})

	lifecycle.AfterStart(func(ctx context.Context) error {
		SetDefault(reg)

		register(ctx, reg)

		cancel := async.GoCtx(func(loopCtx context.Context) error {
			interval := DefaultRegisterInterval
			if cfg.RegisterInterval > 0 {
				interval = cfg.RegisterInterval
			}

			tick := time.NewTicker(interval)
			defer tick.Stop()

			for {
				select {
				case <-tick.C:
					register(loopCtx, reg)
				case <-loopCtx.Done():
					log.Info().Msg("service register cancelled")
					return nil
				}
			}
		})

		lifecycle.BeforeStop(func(stopCtx context.Context) error {
			cancel()
			deregister(stopCtx, reg)
			return nil
		})
		return nil
	})
}

// buildNode 构造当前进程对应的注册节点信息。
func buildNode() *service.Node {
	host := netutil.GetLocalIP()
	port := int(running.GrpcPort.Value())
	return &service.Node{
		Port:    port,
		Version: version.Version(),
		Address: fmt.Sprintf("%s:%d", host, port),
		Id:      running.Project() + "-" + running.Hostname + "-" + running.InstanceID,
	}
}

// buildService 构造注册/撤销用的 Service 对象。
func buildService(reg Registry, includeRegistryMeta bool) *service.Service {
	node := buildNode()
	if includeRegistryMeta {
		node.Metadata = map[string]string{"registry": reg.String()}
	} else {
		node.Metadata = make(map[string]string)
	}
	return &service.Service{
		Name:  running.Project(),
		Nodes: []*service.Node{node},
	}
}

func register(ctx context.Context, reg Registry) {
	s := buildService(reg, true)
	node := s.Nodes[0]

	logutil.OkOrFailed(
		log.GetLogger("service-registry"),
		"register service node",
		func() error {
			err := reg.Register(ctx, s)
			return errors.WrapTags(err, errors.Tags{
				"instance_id": node.Id,
				"service":     running.Project(),
				"registry":    reg.String(),
			})
		},
	)
}

func deregister(ctx context.Context, reg Registry) {
	s := buildService(reg, false)
	node := s.Nodes[0]

	logutil.OkOrFailed(
		log.GetLogger("service-registry"),
		"deregister service node",
		func() error {
			err := reg.Deregister(ctx, s)
			return errors.WrapTags(err, errors.Tags{
				"id":   node.Id,
				"name": running.Project(),
			})
		},
	)
}
