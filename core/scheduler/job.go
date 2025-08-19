package scheduler

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/pubgo/funk/generic"
	"github.com/pubgo/funk/log"
	"github.com/pubgo/funk/try"
	"github.com/pubgo/funk/v2/result"
	"github.com/pubgo/lava/core/metrics"
	"github.com/reugn/go-quartz/quartz"
	"github.com/rs/zerolog"
	"github.com/samber/lo"
)

type jobWrapper struct {
	key  string
	cron string
	dur  time.Duration
	once bool
}

type namedJob struct {
	s    *Scheduler
	name string
	log  log.Logger

	task *jobTask
}

func (t *namedJob) Description() string { return t.name }
func (t *namedJob) Execute(ctx context.Context) (gErr error) {
	start := time.Now()

	defer func() {
		cost := float64(time.Since(start).Milliseconds())
		t.s.metric.Tagged(metrics.Tags{"job_name": t.name}).Gauge("job_cost_ms").Update(cost)

		logger := generic.Ternary(generic.IsNil(gErr), t.log.Info(), t.log.Err(gErr))
		logger.Func(func(e *zerolog.Event) {
			e.Float32("job_cost_ms", float32(cost))
			e.Str("job_name", t.name)
			e.Uint64("runs", t.task.runs.Load())
			e.Msg("exec scheduler job")
		})
	}()

	t.task.runs.Inc()
	metadata := JobMetadata{
		Name:          t.task.config.Name,
		Replace:       lo.FromPtr(t.task.config.Replace),
		MaxRetries:    lo.FromPtr(t.task.config.MaxRetries),
		RetryInterval: lo.FromPtr(t.task.config.RetryInterval),
		Timeout:       lo.FromPtr(t.task.config.Timeout),
		Location:      t.task.config.location,
		PreRunTime:    t.task.trigger.prev,
		NextRunTime:   t.task.trigger.next,
	}

	if t.task.trigger.err != nil {
		return fmt.Errorf("schedule job(%s) error: %w", t.name, t.task.trigger.err)
	}

	return try.Try(func() error {
		ctx, cancel := context.WithTimeout(ctx, lo.FromPtr(t.task.config.Timeout))
		defer cancel()

		t.task.result = t.task.executor.Exec(ctx, t.name, &metadata)
		return t.task.result.GetErr()
	})
}

func registerJob(s *Scheduler, job jobWrapper, fn JobFunc) (r result.Error) {
	s.configMap[job.key] = initConfig(job.key, s.configMap[job.key]).UnwrapErr(&r)
	if r.IsErr() {
		return
	}

	config := s.configMap[job.key]
	trigger := getTrigger(job, config.location).UnwrapErr(&r)
	if r.IsErr() {
		return
	}

	if fn == nil {
		return result.Errorf("schedule job(%s) error: %s", job.key, "fn is nil")
	}

	jobOpt := &quartz.JobDetailOptions{
		MaxRetries:    lo.FromPtr(config.MaxRetries),
		RetryInterval: lo.FromPtr(config.RetryInterval),
		Replace:       lo.FromPtr(config.Replace),
		Suspended:     false,
	}

	return result.ErrOf(s.scheduler.ScheduleJob(
		quartz.NewJobDetailWithOptions(
			&namedJob{s: s, name: job.key, fn: fn, log: s.log, config: config, trigger: trigger},
			quartz.NewJobKey(job.key),
			jobOpt,
		),
		trigger,
	))
}

var _ quartz.Trigger = &triggerImpl{}

func newTrigger(trigger quartz.Trigger) *triggerImpl {
	return &triggerImpl{trigger: trigger}
}

type triggerImpl struct {
	prev    int64
	next    int64
	err     error
	trigger quartz.Trigger
}

func (t *triggerImpl) NextFireTime(prev int64) (next int64, err error) {
	t.prev = prev

	defer func() { t.next, t.err = next, err }()
	return t.trigger.NextFireTime(prev)
}

func (t *triggerImpl) Description() string {
	return t.trigger.Description()
}

func getTrigger(j AddJobSpec, location *time.Location) (r result.Result[*triggerImpl]) {
	if j.Once != nil {
		return r.WithValue(newTrigger(quartz.NewRunOnceTrigger(j.Once.Delay)))
	}

	if j.Cron != nil {
		trigger, err := quartz.NewCronTriggerWithLoc(j.Cron.Expr, location)
		if err != nil {
			return r.WithErrorf("cron-expr:%s, err:%s", j.Cron.Expr, err.Error())
		}
		return r.WithValue(newTrigger(trigger))
	}

	if j.Ticker != nil {
		return r.WithValue(newTrigger(quartz.NewSimpleTrigger(j.Ticker.Dur)))
	}

	return r.WithErrorf("please init Once, Cron or Ticker")
}

func parseJobKey(name string) *quartz.JobKey {
	keys := strings.SplitN(name, quartz.Sep, 2)
	if len(keys) == 1 {
		return quartz.NewJobKey(keys[0])
	}
	return quartz.NewJobKeyWithGroup(keys[0], keys[1])
}
