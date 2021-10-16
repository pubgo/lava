package scheduler

import (
	"time"

	"github.com/pubgo/xerror"
	"github.com/reugn/go-quartz/quartz"

	"github.com/pubgo/lava/logger"
	"github.com/pubgo/lava/logz"
)

var quart = &Scheduler{scheduler: quartz.NewStdScheduler()}

type Scheduler struct {
	scheduler quartz.Scheduler
	key       string
	cron      string
	dur       time.Duration
	once      bool
}

func (s Scheduler) do(fn func(name string)) {
	var trigger = s.getTrigger()
	s.check(s.key, fn, trigger)

	xerror.Panic(s.scheduler.ScheduleJob(nameJob{name: s.key, fn: fn}, trigger))
}

func (s *Scheduler) Once(name string, delay time.Duration, fn func(name string)) {
	logz.Named(Name, 1).Infof("register scheduler(%s) Once(%s)", name, delay)
	Scheduler{scheduler: s.scheduler, dur: delay, key: name, once: true}.do(fn)
}

func (s *Scheduler) Every(name string, dur time.Duration, fn func(name string)) {
	logz.Named(Name, 1).Infof("register scheduler(%s) Every(%s)", name, dur)
	Scheduler{scheduler: s.scheduler, dur: dur, key: name}.do(fn)
}

func (s *Scheduler) Cron(name string, expr string, fn func(name string)) {
	logz.Named(Name, 1).Infof("register scheduler(%s) Cron(%s)", name, expr)
	Scheduler{scheduler: s.scheduler, cron: expr, key: name}.do(fn)
}

func (s *Scheduler) getTrigger() quartz.Trigger {
	if s.once {
		return quartz.NewRunOnceTrigger(s.dur)
	}

	if s.cron != "" {
		return xerror.PanicErr(quartz.NewCronTrigger(s.cron)).(*quartz.CronTrigger)
	}

	if s.dur != 0 {
		return quartz.NewSimpleTrigger(s.dur)
	}

	return nil
}

func (s *Scheduler) check(name string, fn func(name string), trigger quartz.Trigger) {
	xerror.Assert(name == "", "[name] should not be null")
	xerror.Assert(fn == nil, "[fn] should not be nil")
	xerror.Assert(trigger == nil, "please init dur or cron")
}

type nameJob struct {
	name string
	fn   func(name string)
}

func (t nameJob) Description() string { return t.name }
func (t nameJob) Key() int            { return quartz.HashCode(t.Description()) }
func (t nameJob) Execute() {
	var now = time.Now()
	defer func() { logz.Named(Name).Infof("scheduler(%s) trigger ok, duration=>%s", t.name, time.Since(now)) }()

	defer xerror.Resp(func(err xerror.XErr) {
		logz.With(Name, logger.WithErr(err)...).Errorf("scheduler(%s) trigger error, duration=>%s", t.name, time.Since(now))
	})

	t.fn(t.name)
}
