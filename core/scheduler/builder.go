package scheduler

import (
	"context"
	"log/slog"

	"github.com/pubgo/funk/log"
	"github.com/pubgo/funk/log/logfields"
	"github.com/pubgo/funk/v2/result"
	"github.com/pubgo/funk/vars"
	qlog "github.com/reugn/go-quartz/logger"
	"github.com/reugn/go-quartz/quartz"
	"github.com/rs/zerolog"

	"github.com/pubgo/lava/v2/core/lifecycle"
	"github.com/pubgo/lava/v2/core/metrics"
	"github.com/pubgo/lava/v2/core/supervisor"
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
	defer result.Recovery(&gErr)
	configMap := createConfig(configs).
		Log(func(e *zerolog.Event) {
			e.Any("configs", configs)
			e.Any(logfields.Msg, "failed to create config")
		}).
		Must()

	ctx, cancel := context.WithCancel(context.Background())

	slogLogger := qlog.NewSlogLogger(ctx, slog.With(slog.String(logfields.Module, Name)))
	scheduler := result.Wrap(quartz.NewStdScheduler(quartz.WithLogger(slogLogger), quartz.WithJobMetadata())).
		Log(func(e *zerolog.Event) {
			e.Str(logfields.Msg, "failed to create scheduler")
		}).
		Must()

	jobExecutors := make(map[string]JobExecutor)
	for _, executor := range executors {
		regJobExecutor(jobExecutors, executor).Log(func(e *zerolog.Event) {
			e.Str(logfields.Msg, "failed to register job executor")
		}).Must()
	}

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

	vars.Register(vars.UniqueName(Name), func() interface{} {
		return quart.ListJobs()
	})

	return quart, nil
}
