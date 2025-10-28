package scheduler

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/pubgo/funk/v2/log"
	"github.com/pubgo/funk/v2/log/logfields"
	"github.com/pubgo/funk/v2/result"
	"github.com/reugn/go-quartz/quartz"
	"github.com/rs/zerolog"

	"github.com/pubgo/lava/v2/core/metrics"
)

var (
	_ JobManager  = (*Scheduler)(nil)
	_ JobRegistry = (*Scheduler)(nil)
)

type Scheduler struct {
	metric       metrics.Metric
	configMap    map[string]*JobConfig
	scheduler    quartz.Scheduler
	log          log.Logger
	cancel       context.CancelFunc
	ctx          context.Context
	jobExecutors map[string]JobExecutor

	mu sync.Mutex

	jobs sync.Map
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
	defer result.Recovery(&r)

	if spec.Name == "" {
		return r.WithErrorf("job name is empty")
	}

	name := spec.Name
	if _, ok := s.jobs.Load(name); ok {
		return r.WithErrorf("job %s already exists", name)
	}

	executorRes := result.WrapFn(func() (JobExecutor, error) {
		executor := s.jobExecutors[spec.Executor]
		if executor == nil {
			executor = fn
		}

		if executor == nil {
			return nil, fmt.Errorf("schedule job executor is nil, name:%s", name)
		}
		return executor, nil
	})
	executorRes.IfOK(func(executor JobExecutor) { task.executor = executor })
	if executorRes.ThrowErr(&r) {
		return r
	}

	config := initAndMergeConfig(name, s.configMap[name], spec.Config).
		Log(func(e *zerolog.Event) {
			e.Str(logfields.Msg, fmt.Sprintf("failed to init schedule job(%s) config", name))
		}).
		IfOK(func(config *JobConfig) {
			task.spec.Config = config
		}).
		UnwrapOrThrow(&r)
	if r.IsErr() {
		return r
	}

	triggerRes := getTrigger(spec, config.location).
		IfErr(func(err error) {
			log.Err(err).Msgf("failed to get schedule job(%s) trigger", name)
		}).
		IfOK(func(trigger *triggerImpl) {
			task.trigger = trigger
		})
	if triggerRes.ThrowErr(&r) {
		return r
	}

	jobOpt := config.ToJobDetailOptions()
	job := &namedJob{s: s, task: &task, log: s.log}
	jobDetail := quartz.NewJobDetailWithOptions(job, parseJobKey(name), jobOpt)

	if result.Throw(&r, s.scheduler.ScheduleJob(jobDetail, task.trigger)) {
		return r
	}

	s.jobs.Store(name, &task)
	return r
}

func (s *Scheduler) CreateJob(spec JobSpec) (r result.Error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.createJob(spec, nil)
}

func (s *Scheduler) getJob(name string) (r result.Result[*jobTask]) {
	if val, ok := s.jobs.Load(name); !ok {
		return r.WithErrorf("job %s not exists", name)
	} else {
		return r.WithValue(val.(*jobTask))
	}
}

func (s *Scheduler) PatchJob(name string, config *JobConfig) (r result.Error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	job := s.getJob(name).UnwrapOrThrow(&r)
	if r.IsErr() {
		return r
	}

	initAndMergeConfig(name, job.spec.Config, config).
		Log(func(e *zerolog.Event) {
			e.Str(logfields.Msg, fmt.Sprintf("failed to patch schedule job(%s) config", name))
		}).
		IfOK(func(config *JobConfig) {
			job.spec.Config = config
		}).
		ThrowErr(&r)

	return r
}

func (s *Scheduler) PauseJob(name string) (r result.Error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	job := s.getJob(name).UnwrapOrThrow(&r)
	if r.IsErr() {
		return r
	}

	job.status = StatusStop
	return result.ErrOf(s.scheduler.PauseJob(job.jobKey)).IfErr(func(err error) {
		log.Err(err).Msgf("failed to pause schedule job(%s)", name)
	})
}

func (s *Scheduler) ResumeJob(name string) (r result.Error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	job := s.getJob(name).UnwrapOrThrow(&r)
	if r.IsErr() {
		return r
	}

	job.status = StatusRunning
	return result.ErrOf(s.scheduler.ResumeJob(job.jobKey)).
		IfErr(func(err error) {
			log.Err(err).Msgf("failed to resume schedule job(%s)", name)
		})
}

func (s *Scheduler) DeleteJob(name string) (r result.Error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	job := s.getJob(name).UnwrapOrThrow(&r)
	if r.IsErr() {
		return r
	}

	s.jobs.Delete(name)
	return result.ErrOf(s.scheduler.DeleteJob(job.jobKey)).
		IfErr(func(err error) {
			log.Err(err).Msgf("failed to delete schedule job(%s)", name)
		})
}

func (s *Scheduler) ReloadJob(name string) (r result.Error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	job := s.getJob(name).UnwrapOrThrow(&r)
	if r.IsErr() {
		return r
	}

	jobOpt := job.spec.Config.ToJobDetailOptions()
	jobDetail := quartz.NewJobDetailWithOptions(
		&namedJob{s: s, task: job, log: s.log},
		job.jobKey,
		jobOpt,
	)

	jj, _ := s.scheduler.GetScheduledJob(job.jobKey)
	if jj != nil {
		if result.Throw(&r, s.scheduler.DeleteJob(job.jobKey)) {
			return r
		}
	}

	if result.Throw(&r, s.scheduler.ScheduleJob(jobDetail, job.trigger)) {
		return r
	}
	return r
}

func (s *Scheduler) ListJobs() []*Job {
	s.mu.Lock()
	defer s.mu.Unlock()

	var jobs []*Job
	s.jobs.Range(func(key, value any) bool {
		jobs = append(jobs, value.(*jobTask).ToJob())
		return true
	})
	return jobs
}

func (s *Scheduler) GetJob(name string) (r result.Result[*Job]) {
	s.mu.Lock()
	defer s.mu.Unlock()

	job := s.getJob(name).UnwrapOrThrow(&r)
	if r.IsErr() {
		return r
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
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.createJob(JobSpec{Name: name, Once: &OnceJob{Delay: delay}}, fn)
}

func (s *Scheduler) Every(name string, dur time.Duration, fn JobFunc) result.Error {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.createJob(JobSpec{Name: name, Ticker: &TickerJob{Dur: dur}}, fn)
}

func (s *Scheduler) Cron(name, expr string, fn JobFunc) result.Error {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.createJob(JobSpec{Name: name, Cron: &CronJob{Expr: expr}}, fn)
}
