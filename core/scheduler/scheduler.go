package scheduler

import (
	"context"
	"fmt"
	"time"

	"github.com/pubgo/funk/log"
	"github.com/pubgo/funk/v2/result"
	"github.com/pubgo/lava/core/metrics"
	"github.com/reugn/go-quartz/quartz"
	"github.com/rs/zerolog"
	"go.uber.org/atomic"
)

type jobTask struct {
	name     string
	executor JobExecutor
	config   *JobConfig

	Once   *OnceJob
	Ticker *TickerJob
	Cron   *CronJob

	trigger *triggerImpl
	runs    atomic.Uint64

	result result.Result[[]byte]
}

var _ JobManager = (*Scheduler)(nil)
var _ JobRegistry = (*Scheduler)(nil)

type Scheduler struct {
	metric       metrics.Metric
	configMap    map[string]*JobConfig
	scheduler    quartz.Scheduler
	log          log.Logger
	cancel       context.CancelFunc
	ctx          context.Context
	jobs         map[string]*jobTask
	jobExecutors map[string]JobExecutor
}

func regJobExecutor(jobExecutors map[string]JobExecutor, executor JobExecutor) (r result.Error) {
	if executor == nil {
		return result.Errorf("executor is nil")
	}

	if executor.Name() == "" {
		return result.Errorf("executor name is empty")
	}

	if jobExecutors[executor.Name()] != nil {
		return result.Errorf("[job executor] %s already exists", executor.Name())
	}

	jobExecutors[executor.Name()] = executor
	return
}

func (s *Scheduler) PatchJob(name string, config *JobConfig) result.Error {
	//TODO implement me
	panic("implement me")
}

func (s *Scheduler) createJob(spec JobSpec, fn JobFunc) (r result.Error) {
	task := jobTask{Once: spec.Once, Ticker: spec.Ticker, Cron: spec.Cron, name: spec.Name}

	defer func() {
		logFn := func(e *zerolog.Event) {
			e.Str("name", task.name)

			if task.executor != nil {
				e.Str("executor", task.executor.Name())
			}

			if task.config != nil {
				e.Any("config", task.config)
			}
			e.Any("once", task.Once)
			e.Any("ticker", task.Ticker)
			e.Any("cron", task.Cron)
		}
		if r.IsErr() {
			s.log.Err(r.GetErr()).Func(logFn).Msg("failed to register scheduler job")
		} else {
			s.log.Info().Func(logFn).Msg("register scheduler job ok")
		}
	}()
	defer result.RecoveryErr(&r)

	if spec.Name == "" {
		return result.Errorf("job name is empty")
	}

	name := spec.Name
	if s.jobs[name] != nil {
		return result.Errorf("job %s already exists", name)
	}

	result.WrapFn(func() (JobExecutor, error) {
		executor := s.jobExecutors[spec.Executor]
		if executor == nil {
			executor = fn
		}

		if executor == nil {
			return nil, fmt.Errorf("schedule job(%s) executor is nil", name)
		}
		return executor, nil
	}).
		Inspect(func(executor JobExecutor) {
			task.executor = executor
		}).
		UnwrapErr(&r)
	if r.IsErr() {
		return
	}

	config := initConfig(name, s.configMap[name], &spec.Config).
		InspectErr(func(err error) {
			s.log.Err(err).Msgf("failed to init schedule job(%s) config", name)
		}).
		Inspect(func(config *JobConfig) {
			task.config = config
		}).
		UnwrapErr(&r)
	if r.IsErr() {
		return
	}

	trigger := getTrigger(spec, config.location).
		Inspect(func(trigger *triggerImpl) {
			task.trigger = trigger
		}).
		UnwrapErr(&r)
	if r.IsErr() {
		return
	}

	jobOpt := config.ToJobDetailOptions()
	job := &namedJob{s: s, name: name, task: &task, log: s.log}
	jobDetail := quartz.NewJobDetailWithOptions(job, parseJobKey(name), jobOpt)
	if result.CatchErr(&r, s.scheduler.ScheduleJob(jobDetail, trigger)) {
		return
	}

	s.jobs[name] = &task
	return
}

func (s *Scheduler) CreateJob(spec JobSpec) (r result.Error) {
	return s.createJob(spec, nil)
}

func (s *Scheduler) PauseJob(name string) result.Error {
	//TODO implement me
	panic("implement me")
}

func (s *Scheduler) ResumeJob(name string) result.Error {
	//TODO implement me
	panic("implement me")
}

func (s *Scheduler) DeleteJob(name string) result.Error {
	//TODO implement me
	panic("implement me")
}

func (s *Scheduler) ReloadJob(name string) result.Error {
	//TODO implement me
	panic("implement me")
}

func (s *Scheduler) ListJobs() []Job {
	//TODO implement me
	panic("implement me")
}

func (s *Scheduler) GetJob(name string) Job {
	//TODO implement me
	panic("implement me")
}

func (s *Scheduler) String() string {
	return Name
}

func (s *Scheduler) Serve(ctx context.Context) error {
	defer s.stop()
	s.start()

	s.scheduler.Wait(ctx)
	return nil
}

func (s *Scheduler) stop() {
	s.cancel()

	if s.scheduler.IsStarted() {
		s.scheduler.Stop()
	}
}

func (s *Scheduler) start() {
	if s.scheduler.IsStarted() {
		return
	}

	s.scheduler.Start(s.ctx)
}

func (s *Scheduler) Once(name string, delay time.Duration, fn JobFunc) result.Error {
	return s.createJob(JobSpec{Name: name, Once: &OnceJob{Delay: delay}}, fn)
}

func (s *Scheduler) Every(name string, dur time.Duration, fn JobFunc) result.Error {
	return s.createJob(JobSpec{Name: name, Ticker: &TickerJob{Dur: dur}}, fn)
}

func (s *Scheduler) Cron(name, expr string, fn JobFunc) result.Error {
	return s.createJob(JobSpec{Name: name, Cron: &CronJob{Expr: expr}}, fn)
}
