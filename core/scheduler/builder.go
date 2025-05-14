package scheduler

import (
	"context"
	"fmt"

	"github.com/pubgo/funk/log"
	"github.com/reugn/go-quartz/quartz"

	"github.com/pubgo/lava/core/lifecycle"
	"github.com/pubgo/lava/core/metrics"
)

const Name = "scheduler"

func New(m lifecycle.Lifecycle, log log.Logger, opts []*Config, routers []Register, metric metrics.Metric) *Scheduler {
	config := make(map[string]JobSetting)
	if len(opts) > 0 && opts[0] != nil {
		for _, setting := range opts[0].JobSettings {
			if _, ok := config[setting.Name]; ok {
				panic(fmt.Sprintf("schedule job(%s) exists", setting.Name))
			}

			config[setting.Name] = setting
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	quart := &Scheduler{
		metric:    metric,
		config:    config,
		scheduler: quartz.NewStdScheduler(),
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

	return quart
}
