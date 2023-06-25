package scheduler

import (
	"context"
	"time"

	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/errors"
	"github.com/pubgo/funk/generic"
	"github.com/pubgo/funk/log"
	"github.com/pubgo/funk/result"
	"github.com/pubgo/funk/try"
	"github.com/reugn/go-quartz/quartz"
)

type job struct {
	key  string
	cron string
	dur  time.Duration
	once bool
}

type Scheduler struct {
	config    map[string]JobSetting
	scheduler quartz.Scheduler
	log       log.Logger
	cancel    context.CancelFunc
	ctx       context.Context
}

func (s *Scheduler) stop() {
	s.cancel()
	s.scheduler.Stop()
}

func (s *Scheduler) start() {
	if s.scheduler.IsStarted() {
		return
	}

	s.scheduler.Start()
}

func (s *Scheduler) Once(name string, delay time.Duration, fn func(ctx context.Context, name string) error) {
	s.log.WithCallerSkip(1).Info().
		Str("name", name).
		Str("delay", delay.String()).
		Msg("register once scheduler")
	do(s, job{dur: delay, key: name, once: true}, fn)
}

func (s *Scheduler) Every(name string, dur time.Duration, fn func(ctx context.Context, name string) error) {
	s.log.WithCallerSkip(1).Info().
		Str("name", name).
		Str("dur", dur.String()).
		Msg("register periodic scheduler")
	do(s, job{dur: dur, key: name}, fn)
}

func (s *Scheduler) Cron(name string, expr string, fn func(ctx context.Context, name string) error) {
	s.log.WithCallerSkip(1).Info().
		Str("name", name).
		Str("expr", expr).
		Msg("register cron scheduler")
	do(s, job{cron: expr, key: name}, fn)
}

func getTrigger(j job) quartz.Trigger {
	if j.once {
		return quartz.NewRunOnceTrigger(j.dur)
	}

	if j.cron != "" {
		r := result.Wrap(quartz.NewCronTrigger(j.cron))
		return r.Unwrap(func(err error) error {
			return errors.WrapKV(err, "cron-expr", j.cron)
		})
	}

	if j.dur != 0 {
		return quartz.NewSimpleTrigger(j.dur)
	}

	return nil
}

func do(s *Scheduler, job job, fn func(ctx context.Context, name string) error) {
	trigger := getTrigger(job)
	assert.If(job.key == "", "[name] should not be null")
	assert.If(fn == nil, "[fn] should not be nil")
	assert.If(trigger == nil, "please init dur or cron")
	assert.Must(s.scheduler.ScheduleJob(namedJob{s: s, name: job.key, fn: fn, log: s.log}, trigger))
}

type namedJob struct {
	s    *Scheduler
	name string
	fn   func(ctx context.Context, name string) error
	log  log.Logger
}

func (t namedJob) Description() string { return t.name }
func (t namedJob) Key() int            { return quartz.HashCode(t.Description()) }
func (t namedJob) Execute() {
	start := time.Now()
	err := try.Try(func() error {
		ctx, cancel := context.WithCancel(t.s.ctx)
		defer cancel()

		return t.fn(ctx, t.name)
	})

	logger := generic.Ternary(generic.IsNil(err), t.log.Info(), t.log.Err(err))
	logger.
		Float32("job-cost-ms", float32(time.Since(start).Microseconds())/1000).
		Str("job-name", t.name).
		Msg("scheduler job execution")
}
