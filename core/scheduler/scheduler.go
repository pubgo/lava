package scheduler

import (
	"context"
	"fmt"
	"time"

	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/errors"
	"github.com/pubgo/funk/log"
	"github.com/pubgo/funk/stack"
	"github.com/pubgo/funk/v2/result"
	"github.com/pubgo/lava/core/metrics"
	"github.com/reugn/go-quartz/quartz"
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

func (s *Scheduler) Patch(name string, config JobConfig) error {
	//TODO implement me
	panic("implement me")
}

func (s *Scheduler) Create(spec AddJobSpec) (r result.Error) {
	if spec.Name == "" {
		return result.Errorf("job name is empty")
	}

	name := spec.Name
	if s.jobs[name] != nil {
		return result.Errorf("job %s already exists", name)
	}

	config := initConfig(name, s.configMap[name], &spec.Config).
		InspectErr(func(err error) {
			s.log.Err(err).Msgf("failed to init schedule job(%s) config", name)
		}).
		UnwrapErr(&r)
	if r.IsErr() {
		return
	}

	executor := GetJobExecutor(spec.Executor)
	if executor == nil {
		return result.ErrorOf("schedule job(%s) error: %s", name, "executor is nil")
	}

	trigger := getTrigger(spec, config.location).UnwrapErr(&r)
	if r.IsErr() {
		return
	}

	task := &jobTask{
		name:     name,
		executor: executor,
		config:   config,
		trigger:  trigger,

		Once:   spec.Once,
		Ticker: spec.Ticker,
		Cron:   spec.Cron,
	}
	s.jobs[name] = task

	jobOpt := config.ToJobDetailOptions()
	job := &namedJob{s: s, name: name, task: task, log: s.log}
	return result.ErrOf(s.scheduler.ScheduleJob(
		quartz.NewJobDetailWithOptions(job, parseJobKey(name), jobOpt),
		trigger,
	))
}

func (s *Scheduler) Pause(name string) error {
	//TODO implement me
	panic("implement me")
}

func (s *Scheduler) Resume(name string) error {
	//TODO implement me
	panic("implement me")
}

func (s *Scheduler) Delete(name string) error {
	//TODO implement me
	panic("implement me")
}

func (s *Scheduler) Reload(name string) error {
	//TODO implement me
	panic("implement me")
}

func (s *Scheduler) List() []Job {
	//TODO implement me
	panic("implement me")
}

func (s *Scheduler) Get(name string) Job {
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

func (s *Scheduler) checkJobExists(name string, fn JobFunc) error {
	if s.jobs[name] != nil {
		return &errors.Err{
			Msg:    fmt.Sprintf("job %s exists", name),
			Detail: stack.CallerWithFunc(s.jobs[name]).String(),
		}
	}

	s.jobs[name] = fn
	return nil
}

func (s *Scheduler) Once(name string, delay time.Duration, fn JobFunc) {
	assert.Must(s.checkJobExists(name, fn))

	s.log.WithCallerSkip(1).Info().
		Str("name", name).
		Str("delay", delay.String()).
		Msg("register once scheduler")
	registerJob(s, jobWrapper{dur: delay, key: name, once: true}, fn)
}

func (s *Scheduler) Every(name string, dur time.Duration, fn JobFunc) {
	assert.Must(s.checkJobExists(name, fn))

	s.log.WithCallerSkip(1).Info().
		Str("name", name).
		Str("dur", dur.String()).
		Msg("register periodic scheduler")
	registerJob(s, jobWrapper{dur: dur, key: name}, fn)
}

func (s *Scheduler) Cron(name, expr string, fn JobFunc) {
	assert.Must(s.checkJobExists(name, fn))

	s.log.WithCallerSkip(1).Info().
		Str("name", name).
		Str("expr", expr).
		Msg("register cron scheduler")
	registerJob(s, jobWrapper{cron: expr, key: name}, fn)
}
