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

func New(m lifecycle.Lifecycle, logger log.Logger, metric metrics.Metric, configs []*Config, routers []JobRegister, executors []JobExecutor) (_ *Scheduler, gErr error) {
	defer result.RecoveryErr(&gErr)

	configMap := result.Wrap(createConfig(configs)).Must(func(e *zerolog.Event) {
		e.Any("configs", configs)
		e.Any(logfields.Msg, "failed to create config")
	})

	ctx, cancel := context.WithCancel(context.Background())

	slogLogger := qlog.NewSlogLogger(ctx, slog.With(slog.String(logfields.Module, Name)))
	scheduler := result.Wrap(quartz.NewStdScheduler(quartz.WithLogger(slogLogger), quartz.WithJobMetadata())).
		Must(func(e *zerolog.Event) {
			e.Str(logfields.Msg, "failed to create scheduler")
		})

	jobExecutors := make(map[string]JobExecutor)
	for _, executor := range executors {
		regJobExecutor(jobExecutors, executor).Must(func(e *zerolog.Event) {
			e.Str(logfields.Msg, "failed to register job executor")
		})
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

	m.AfterStart(func(ctx context.Context) error { return lifecycle.WrapNoCtxErr(quart.start)(ctx) })
	m.BeforeStop(func(ctx context.Context) error { return lifecycle.WrapNoCtxErr(quart.stop)(ctx) })

	vars.Register(vars.UniqueName(Name), func() any { return quart.ListJobs() })

	return quart, nil
}
