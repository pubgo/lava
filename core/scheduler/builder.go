package scheduler

import (
	"context"
	"fmt"
	"github.com/pubgo/funk/log"
	"github.com/pubgo/funk/v2/result"
	"github.com/reugn/go-quartz/quartz"

	"github.com/pubgo/lava/core/lifecycle"
	"github.com/pubgo/lava/core/metrics"
	"github.com/pubgo/lava/core/supervisor"
)

const Name = "scheduler"

type Params struct {
	M       lifecycle.Lifecycle
	Log     log.Logger
	Configs []*Config
	Routers []Register
	Metric  metrics.Metric
}

func NewService(params Params) (supervisor.Service, error) {
	s, err := New(
		params.M,
		params.Log,
		params.Configs,
		params.Routers,
		params.Metric,
	)
	if err != nil {
		return nil, err
	}
	return supervisor.NewService(Name, s.Serve), err
}

func New(m lifecycle.Lifecycle, log log.Logger, opts []*Config, routers []Register, metric metrics.Metric) (_ *Scheduler, gErr error) {
	config := createConfig(opts).Unwrap(&gErr)
	if gErr != nil {
		return nil, fmt.Errorf("failed to create config, err:%w", gErr)
	}

	scheduler := result.Wrap(quartz.NewStdScheduler()).Unwrap(&gErr)
	if gErr != nil {
		return nil, fmt.Errorf("failed to create scheduler, err:%w", gErr)
	}

	ctx, cancel := context.WithCancel(context.Background())
	quart := &Scheduler{
		metric:    metric,
		configMap: config,
		scheduler: scheduler,
		log:       log.WithName(Name),
		ctx:       ctx,
		cancel:    cancel,
		jobs:      make(map[string]JobFunc),
	}

	quart.start()
	m.BeforeStop(lifecycle.WrapNoCtxErr(quart.stop))

	for _, r := range routers {
		r.RegisterCrontabScheduler(quart)
	}

	return quart, nil
}
