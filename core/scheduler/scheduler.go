package scheduler

import (
	"time"

	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/errors"
	"github.com/pubgo/funk/generic"
	"github.com/pubgo/funk/log"
	"github.com/pubgo/funk/result"
	"github.com/pubgo/funk/try"
	"github.com/reugn/go-quartz/quartz"
)

type Scheduler struct {
	scheduler quartz.Scheduler
	key       string
	cron      string
	dur       time.Duration
	once      bool
	log       log.Logger
}

func (s *Scheduler) Stop() {
	s.scheduler.Stop()
}

func (s *Scheduler) Start() {
	if s.scheduler.IsStarted() {
		return
	}
	s.scheduler.Start()
}

func (s *Scheduler) Once(name string, delay time.Duration, fn func(name string)) {
	s.log.WithCallerSkip(1).Info().
		Str("name", name).
		Str("delay", delay.String()).
		Msg("register once scheduler")
	do(&Scheduler{scheduler: s.scheduler, dur: delay, key: name, once: true, log: s.log}, fn)
}

func (s *Scheduler) Every(name string, dur time.Duration, fn func(name string)) {
	s.log.WithCallerSkip(1).Info().
		Str("name", name).
		Str("dur", dur.String()).
		Msg("register periodic scheduler")
	do(&Scheduler{scheduler: s.scheduler, dur: dur, key: name, log: s.log}, fn)
}

func (s *Scheduler) Cron(name string, expr string, fn func(name string)) {
	s.log.WithCallerSkip(1).Info().
		Str("name", name).
		Str("expr", expr).
		Msg("register cron scheduler")
	do(&Scheduler{scheduler: s.scheduler, cron: expr, key: name, log: s.log}, fn)
}

func getTrigger(once bool, cron string, dur time.Duration) quartz.Trigger {
	if once {
		return quartz.NewRunOnceTrigger(dur)
	}

	if cron != "" {
		r := result.Wrap(quartz.NewCronTrigger(cron))
		return r.Unwrap(func(err error) error {
			return errors.WrapKV(err, "cron-expr", cron)
		})
	}

	if dur != 0 {
		return quartz.NewSimpleTrigger(dur)
	}

	return nil
}

func do(s *Scheduler, fn func(name string)) {
	trigger := getTrigger(s.once, s.cron, s.dur)
	assert.If(s.key == "", "[name] should not be null")
	assert.If(fn == nil, "[fn] should not be nil")
	assert.If(trigger == nil, "please init dur or cron")
	assert.Must(s.scheduler.ScheduleJob(namedJob{name: s.key, fn: fn, log: s.log}, trigger))
}

type namedJob struct {
	name string
	fn   func(name string)
	log  log.Logger
}

func (t namedJob) Description() string { return t.name }
func (t namedJob) Key() int            { return quartz.HashCode(t.Description()) }
func (t namedJob) Execute() {
	evt := log.NewEvent().Str("job-name", t.name)
	s := time.Now()
	err := try.Try(func() error { t.fn(t.name); return nil })
	evt.Int64("job-cost-ms", time.Since(s).Milliseconds())
	if generic.IsNil(err) {
		log.Info().Func(log.WithEvent(evt)).Msg("scheduler trigger")
	} else {
		log.Err(err).Func(log.WithEvent(evt)).Msg("scheduler trigger")
	}
}
