package scheduler

import (
	"context"
	"fmt"
	"time"

	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/errors"
	"github.com/pubgo/funk/generic"
	"github.com/pubgo/funk/log"
	"github.com/pubgo/funk/stack"
	"github.com/pubgo/funk/try"
	"github.com/reugn/go-quartz/quartz"

	"github.com/pubgo/lava/core/metrics"
)

type job struct {
	key  string
	cron string
	dur  time.Duration
	once bool
}

type Scheduler struct {
	metric    metrics.Metric
	config    map[string]JobSetting
	scheduler quartz.Scheduler
	log       log.Logger
	cancel    context.CancelFunc
	ctx       context.Context
	jobs      map[string]JobFunc
}

func (s *Scheduler) stop() {
	s.cancel()
	s.scheduler.Stop()
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
	do(s, job{dur: delay, key: name, once: true}, fn)
}

func (s *Scheduler) Every(name string, dur time.Duration, fn JobFunc) {
	assert.Must(s.checkJobExists(name, fn))

	s.log.WithCallerSkip(1).Info().
		Str("name", name).
		Str("dur", dur.String()).
		Msg("register periodic scheduler")
	do(s, job{dur: dur, key: name}, fn)
}

func (s *Scheduler) Cron(name, expr string, fn JobFunc) {
	assert.Must(s.checkJobExists(name, fn))

	s.log.WithCallerSkip(1).Info().
		Str("name", name).
		Str("expr", expr).
		Msg("register cron scheduler")
	do(s, job{cron: expr, key: name}, fn)
}

type namedJob struct {
	s    *Scheduler
	name string
	fn   JobFunc
	log  log.Logger
}

func (t namedJob) Description() string { return t.name }
func (t namedJob) Execute(ctx context.Context) error {
	start := time.Now()
	err := try.Try(func() error {
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		return t.fn(ctx, t.name)
	})

	t.s.metric.Tagged(metrics.Tags{"job_name": t.name}).Gauge("scheduler_job_cost").Update(float64(time.Since(start).Microseconds()) / 1000)

	logger := generic.Ternary(generic.IsNil(err), t.log.Info(), t.log.Err(err))
	logger.
		Float32("job-cost-ms", float32(time.Since(start).Microseconds())/1000).
		Str("job-name", t.name).
		Msg("scheduler job execution")

	return err
}
