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
)

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

func (s *Scheduler) createJob(spec JobSpec, fn JobFunc) (r result.Error) {
	task := jobTask{
		spec:   &spec,
		jobKey: parseJobKey(spec.Name),
		status: StatusRunning,
	}

	defer func() {
		logFn := func(e *zerolog.Event) {
			e.Str("name", task.spec.Name)

			if task.executor != nil {
				e.Str("executor", task.executor.Name())
			}

			if task.spec.Config != nil {
				e.Any("config", task.spec.Config)
			}
			e.Any("once", task.spec.Once)
			e.Any("ticker", task.spec.Ticker)
			e.Any("cron", task.spec.Cron)
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

	config := initAndMergeConfig(name, s.configMap[name], spec.Config).
		InspectErr(func(err error) {
			s.log.Err(err).Msgf("failed to init schedule job(%s) config", name)
		}).
		Inspect(func(config *JobConfig) {
			task.spec.Config = config
		}).
		UnwrapErr(&r)
	if r.IsErr() {
		return
	}

	trigger := getTrigger(spec, config.location).
		InspectErr(func(err error) {
			log.Err(err).Msgf("failed to get schedule job(%s) trigger", name)
		}).
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

func (s *Scheduler) getJob(name string) (r result.Result[*jobTask]) {
	if s.jobs[name] == nil {
		return r.WithErrorf("job %s not exists", name)
	}
	return r.WithValue(s.jobs[name])
}

func (s *Scheduler) PatchJob(name string, config *JobConfig) (r result.Error) {
	job := s.getJob(name).UnwrapErr(&r)
	if r.IsErr() {
		return
	}

	initAndMergeConfig(name, job.spec.Config, config).
		InspectErr(func(err error) {
			s.log.Err(err).Msgf("failed to patch schedule job(%s) config", name)
		}).
		Inspect(func(config *JobConfig) {
			job.spec.Config = config
		}).
		CatchErr(&r)

	return
}

func (s *Scheduler) PauseJob(name string) (r result.Error) {
	job := s.getJob(name).UnwrapErr(&r)
	if r.IsErr() {
		return
	}

	job.status = StatusStop
	return result.ErrOf(s.scheduler.PauseJob(job.jobKey)).
		InspectErr(func(err error) {
			log.Err(err).Msgf("failed to pause schedule job(%s)", name)
		})
}

func (s *Scheduler) ResumeJob(name string) (r result.Error) {
	job := s.getJob(name).UnwrapErr(&r)
	if r.IsErr() {
		return
	}

	job.status = StatusRunning
	return result.ErrOf(s.scheduler.ResumeJob(job.jobKey)).
		InspectErr(func(err error) {
			log.Err(err).Msgf("failed to resume schedule job(%s)", name)
		})
}

func (s *Scheduler) DeleteJob(name string) (r result.Error) {
	job := s.getJob(name).UnwrapErr(&r)
	if r.IsErr() {
		return
	}

	delete(s.jobs, name)
	return result.ErrOf(s.scheduler.DeleteJob(job.jobKey)).
		InspectErr(func(err error) {
			log.Err(err).Msgf("failed to delete schedule job(%s)", name)
		})
}

func (s *Scheduler) ReloadJob(name string) (r result.Error) {
	job := s.getJob(name).UnwrapErr(&r)
	if r.IsErr() {
		return
	}

	jobOpt := job.spec.Config.ToJobDetailOptions()
	jobDetail := quartz.NewJobDetailWithOptions(
		&namedJob{s: s, name: name, task: job, log: s.log},
		job.jobKey,
		jobOpt,
	)
	if result.CatchErr(&r, s.scheduler.ScheduleJob(jobDetail, job.trigger)) {
		return
	}
	return
}

func (s *Scheduler) ListJobs() []*Job {
	var jobs []*Job
	for _, job := range s.jobs {
		jobs = append(jobs, job.ToJob())
	}
	return jobs
}

func (s *Scheduler) GetJob(name string) (r result.Result[*Job]) {
	job := s.getJob(name).UnwrapErr(&r)
	if r.IsErr() {
		return
	}

	return r.WithValue(job.ToJob())
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
