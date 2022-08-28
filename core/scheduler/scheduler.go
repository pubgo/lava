package scheduler

import (
	"github.com/pubgo/funk/result"
	"time"

	"github.com/pubgo/funk/assert"
	"github.com/reugn/go-quartz/quartz"
	"go.uber.org/zap"

	"github.com/pubgo/lava/logging"
	"github.com/pubgo/lava/logging/logutil"
	"github.com/pubgo/lava/pkg/utils"
)

func New(log *logging.Logger) *Scheduler {
	return &Scheduler{
		scheduler: quartz.NewStdScheduler(),
		log:       logging.ModuleLog(log, Name),
	}
}

type Scheduler struct {
	scheduler quartz.Scheduler
	key       string
	cron      string
	dur       time.Duration
	once      bool
	log       *logging.ModuleLogger
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
	s.log.Depth(1).Info("register once scheduler", zap.String("name", name), zap.String("delay", delay.String()))
	do(&Scheduler{scheduler: s.scheduler, dur: delay, key: name, once: true, log: s.log}, fn)
}

func (s *Scheduler) Every(name string, dur time.Duration, fn func(name string)) {
	s.log.Depth(1).Info("register every scheduler", zap.String("name", name), zap.String("dur", dur.String()))
	do(&Scheduler{scheduler: s.scheduler, dur: dur, key: name, log: s.log}, fn)
}

func (s *Scheduler) Cron(name string, expr string, fn func(name string)) {
	s.log.Depth(1).Info("register cron scheduler", zap.String("name", name), zap.String("expr", expr))
	do(&Scheduler{scheduler: s.scheduler, cron: expr, key: name, log: s.log}, fn)
}

func (s *Scheduler) getTrigger() quartz.Trigger {
	if s.once {
		return quartz.NewRunOnceTrigger(s.dur)
	}

	if s.cron != "" {
		return assert.Must1(quartz.NewCronTrigger(s.cron))
	}

	if s.dur != 0 {
		return quartz.NewSimpleTrigger(s.dur)
	}

	return nil
}

func do(s *Scheduler, fn func(name string)) {
	var trigger = s.getTrigger()
	assert.If(s.key == "", "[name] should not be null")
	assert.If(fn == nil, "[fn] should not be nil")
	assert.If(trigger == nil, "please init dur or cron")
	assert.Must(s.scheduler.ScheduleJob(namedJob{name: s.key, fn: fn, log: s.log}, trigger))
}

type namedJob struct {
	name string
	fn   func(name string)
	log  *logging.ModuleLogger
}

func (t namedJob) Description() string { return t.name }
func (t namedJob) Key() int            { return quartz.HashCode(t.Description()) }
func (t namedJob) Execute() {
	var dur, err = utils.Cost(func() { t.fn(t.name) })
	logutil.LogOrErr(t.log.L(), "scheduler trigger",
		func() result.Error { return result.WithErr(err) },
		zap.String("job-name", t.name),
		zap.Int64("job-cost-ms", dur.Milliseconds()))
}
