package scheduler

import (
	"context"

	"github.com/pubgo/funk/log"
	"github.com/pubgo/funk/v2/result"
	"github.com/pubgo/funk/vars"
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

type ResponseParams struct {
	Service supervisor.Service
	Manager JobManager
}

func NewService(params Params) (ResponseParams, error) {
	s, err := New(
		params.M,
		params.Log,
		params.Metric,
		params.Configs,
		params.Routers,
		params.Executors,
	)
	if err != nil {
		return ResponseParams{}, err
	}

	return ResponseParams{
		Service: supervisor.NewService(Name, s.Serve),
		Manager: s,
	}, nil
}

func New(m lifecycle.Lifecycle, logger log.Logger, metric metrics.Metric, configs []*Config, routers []JobRegister, executors []JobExecutor) (_ *Scheduler, gErr error) {
	configMap := createConfig(configs)
	configMap.InspectErr(func(err error) {
		log.Err(err).Any("configs", configs).Msg("failed to create config")
	})
	if configMap.Catch(&gErr) {
		return
	}

	scheduler := result.WrapFn(func() (quartz.Scheduler, error) {
		return quartz.NewStdScheduler(
			quartz.WithLogger(qlog.NewSimpleLogger(schedulerLog, qlog.LevelInfo)),
			quartz.WithJobMetadata(),
		)
	})
	scheduler.InspectErr(func(err error) {
		log.Err(err).Msg("failed to create scheduler")
	})
	if scheduler.Catch(&gErr) {
		return
	}

	jobExecutors := make(map[string]JobExecutor)
	for _, executor := range executors {
		regRes := regJobExecutor(jobExecutors, executor)
		regRes.InspectErr(func(err error) {
			log.Err(err).Msg("failed to register job executor")
		})
		if regRes.Catch(&gErr) {
			return
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	quart := &Scheduler{
		metric:       metric,
		configMap:    configMap.Must(),
		scheduler:    scheduler.Must(),
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

	vars.Register(Name, func() interface{} {
		return quart.ListJobs()
	})

	return quart, nil
}
