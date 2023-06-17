package registry

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/async"
	"github.com/pubgo/funk/errors"
	"github.com/pubgo/funk/log"
	"github.com/pubgo/funk/runmode"
	"github.com/pubgo/funk/version"

	"github.com/pubgo/lava/core/lifecycle"
	"github.com/pubgo/lava/core/service"
	"github.com/pubgo/lava/pkg/netutil"
)

func New(c *Config, lifecycle lifecycle.Lifecycle, regs map[string]Registry) {
	cfg := DefaultCfg()

	// 配置解析
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

	// 服务注册
	lifecycle.AfterStart(func() {
		SetDefault(reg)

		register(reg)

		cancel := async.GoCtx(func(ctx context.Context) error {
			interval := DefaultRegisterInterval

			if cfg.RegisterInterval > time.Duration(0) {
				interval = cfg.RegisterInterval
			}

			tick := time.NewTicker(interval)
			defer tick.Stop()

			for {
				select {
				case <-tick.C:
					register(reg)
				case <-ctx.Done():
					log.Info().Msg("service register cancelled")
					return nil
				}
			}
		})

		// 服务撤销
		lifecycle.BeforeStop(func() {
			cancel()
			deregister(reg)
		})
	})
}

func register(reg Registry) {
	// parse address for host, port
	var advt, host string
	port := runmode.GrpcPort

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
	node := &service.Node{
		Port:     port,
		Version:  version.Version(),
		Address:  fmt.Sprintf("%s:%d", host, port),
		Id:       runmode.Project + "-" + runmode.Hostname + "-" + runmode.InstanceID,
		Metadata: map[string]string{"registry": reg.String()},
	}

	s := &service.Service{
		Name:  runmode.Project,
		Nodes: []*service.Node{node},
	}

	logutil.OkOrFailed(
		log.GetLogger("service-registry"),
		"register service node",
		func() error {
			err := reg.Register(context.Background(), s)
			return errors.WrapTags(err, errors.Tags{
				"instance_id": node.Id,
				"service":     runmode.Project,
				"registry":    Default().String(),
			})
		},
	)
}

func deregister(reg Registry) {
	var advt, host string
	port := runmode.GrpcPort

	parts := strings.Split(advt, ":")
	if len(parts) > 1 {
		host = strings.Join(parts[:len(parts)-1], ":")
		port, _ = strconv.Atoi(parts[len(parts)-1])
	} else {
		host = parts[0]
	}

	// register service
	node := &service.Node{
		Port:     port,
		Address:  fmt.Sprintf("%s:%d", host, port),
		Id:       runmode.Project + "-" + runmode.Hostname + "-" + runmode.InstanceID,
		Metadata: make(map[string]string),
	}

	s := &service.Service{
		Name:  runmode.Project,
		Nodes: []*service.Node{node},
	}

	logutil.OkOrFailed(
		log.GetLogger("service-registry"),
		"deregister service node",
		func() error {
			err := reg.Deregister(context.Background(), s)
			return errors.WrapTags(err, errors.Tags{"id": node.Id, "name": runmode.Project})
		},
	)
}
