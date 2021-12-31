package scheduler

import (
	"time"

	"github.com/pubgo/xerror"
	"github.com/reugn/go-quartz/quartz"
	"go.uber.org/zap"

	"github.com/pubgo/lava/logz"
	"github.com/pubgo/lava/pkg/lavax"
	"github.com/pubgo/lava/resource"
)

var quart = &Scheduler{scheduler: quartz.NewStdScheduler()}
var logs = logz.Component(Name)

var _ resource.Resource = (*Scheduler)(nil)

type Scheduler struct {
	scheduler quartz.Scheduler
	key       string
	cron      string
	dur       time.Duration
	once      bool
}

func (s Scheduler) Close() error                 { return nil }
func (s Scheduler) UpdateResObj(val interface{}) {}
func (s Scheduler) Kind() string                 { return Name }

func (s Scheduler) do(fn func(name string)) {
	var trigger = s.getTrigger()
	s.check(s.key, fn, trigger)

	xerror.Panic(s.scheduler.ScheduleJob(nameJob{name: s.key, fn: fn}, trigger))
}

func (s *Scheduler) Once(name string, delay time.Duration, fn func(name string)) {
	logs.DepthS(1).Infof("register scheduler(%s) Once(%s)", name, delay)
	Scheduler{scheduler: s.scheduler, dur: delay, key: name, once: true}.do(fn)
}

func (s *Scheduler) Every(name string, dur time.Duration, fn func(name string)) {
	logs.DepthS(1).Infof("register scheduler(%s) Every(%s)", name, dur)
	Scheduler{scheduler: s.scheduler, dur: dur, key: name}.do(fn)
}

func (s *Scheduler) Cron(name string, expr string, fn func(name string)) {
	logs.DepthS(1).Infof("register scheduler(%s) Cron(%s)", name, expr)
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
	var dur, err = lavax.Cost(func() { t.fn(t.name) })
	logs.Logs("scheduler trigger",
		func() error { return err },
		zap.String("job-name", t.name),
		zap.Int64("job-cost", dur.Microseconds()),
	)
}
