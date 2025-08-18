package scheduler

import (
	"context"
	"fmt"
	"time"

	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/errors"
	"github.com/pubgo/funk/log"
	"github.com/pubgo/funk/stack"
	"github.com/reugn/go-quartz/quartz"

	"github.com/pubgo/lava/core/metrics"
)

type Scheduler struct {
	metric    metrics.Metric
	configMap map[string]*JobSetting
	scheduler quartz.Scheduler
	log       log.Logger
	cancel    context.CancelFunc
	ctx       context.Context
	jobs      map[string]JobFunc
}

func (s *Scheduler) String() string {
	return Name
}

func (s *Scheduler) Serve(ctx context.Context) error {
	s.start()
	defer s.stop()

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
