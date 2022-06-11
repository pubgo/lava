package scheduler

import (
	"time"

	"github.com/pubgo/xerror"
	"github.com/reugn/go-quartz/quartz"
	"go.uber.org/zap"

	"github.com/pubgo/lava/internal/pkg/utils"
	"github.com/pubgo/lava/logging"
	"github.com/pubgo/lava/logging/logutil"
)

var quart = &Scheduler{scheduler: quartz.NewStdScheduler()}
var logs = logging.Component(Name)

type Scheduler struct {
	scheduler quartz.Scheduler
	key       string
	cron      string
	dur       time.Duration
	once      bool
	log       *logging.Logger
}

func (s Scheduler) do(fn func(name string)) {
	var trigger = s.getTrigger()
	s.check(s.key, fn, trigger)

	xerror.Panic(s.scheduler.ScheduleJob(nameJob{name: s.key, fn: fn}, trigger))
}

func (s *Scheduler) Once(name string, delay time.Duration, fn func(name string)) {
	logs.Depth(1).Info("register once scheduler", zap.String("name", name), zap.String("delay", delay.String()))
	Scheduler{scheduler: s.scheduler, dur: delay, key: name, once: true}.do(fn)
}

func (s *Scheduler) Every(name string, dur time.Duration, fn func(name string)) {
	logs.Depth(1).Info("register every scheduler", zap.String("name", name), zap.String("dur", dur.String()))
	Scheduler{scheduler: s.scheduler, dur: dur, key: name}.do(fn)
}

func (s *Scheduler) Cron(name string, expr string, fn func(name string)) {
	logs.Depth(1).Info("register cron scheduler", zap.String("name", name), zap.String("expr", expr))
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
	var dur, err = utils.Cost(func() { t.fn(t.name) })
	logutil.LogOrErr(logs.L(), "scheduler trigger",
		func() error { return err },
		zap.String("job-name", t.name),
		zap.Int64("job-cost", dur.Microseconds()))
}
