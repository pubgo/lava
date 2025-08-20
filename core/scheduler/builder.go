package scheduler

import (
	"context"

	"github.com/pubgo/funk/log"
	"github.com/pubgo/funk/v2/result"
	qlog "github.com/reugn/go-quartz/logger"
	"github.com/reugn/go-quartz/quartz"

	"github.com/pubgo/lava/core/lifecycle"
	"github.com/pubgo/lava/core/metrics"
	"github.com/pubgo/lava/core/supervisor"
)

const Name = "scheduler"

type Params struct {
	M         lifecycle.Lifecycle
	Log       log.Logger
	Configs   []*Config
	Routers   []JobRegister
	Metric    metrics.Metric
	Executors []JobExecutor
}

func NewService(params Params) (supervisor.Service, error) {
	s, err := New(
		params.M,
		params.Log,
		params.Metric,
		params.Configs,
		params.Routers,
		params.Executors,
	)
	if err != nil {
		return nil, err
	}
	return supervisor.NewService(Name, s.Serve), err
}

func New(m lifecycle.Lifecycle, logger log.Logger, metric metrics.Metric, configs []*Config, routers []JobRegister, executors []JobExecutor) (_ *Scheduler, gErr error) {
	configMap := createConfig(configs).
		InspectErr(func(err error) {
			log.Err(err).Any("configs", configs).Msg("failed to create config")
		}).
		Unwrap(&gErr)
	if gErr != nil {
		return
	}

	scheduler := result.Wrap(quartz.NewStdScheduler(quartz.WithLogger(qlog.NewSimpleLogger(schedulerLog, qlog.LevelDebug)))).
		InspectErr(func(err error) {
			log.Err(err).Msg("failed to create scheduler")
		}).
		Unwrap(&gErr)
	if gErr != nil {
		return
	}

	jobExecutors := make(map[string]JobExecutor)
	for _, executor := range executors {
		regJobExecutor(jobExecutors, executor).
			InspectErr(func(err error) {
				log.Err(err).Msg("failed to register job executor")
			}).
			Catch(&gErr)
		if gErr != nil {
			return
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	quart := &Scheduler{
		metric:       metric,
		configMap:    configMap,
		scheduler:    scheduler,
		log:          logger.WithName(Name),
		ctx:          ctx,
		cancel:       cancel,
		jobs:         make(map[string]*jobTask),
		jobExecutors: jobExecutors,
	}

	for _, r := range routers {
		r.RegisterSchedulerJob(quart)
	}

	quart.start()
	m.BeforeStop(lifecycle.WrapNoCtxErr(quart.stop))

	return quart, nil
}
