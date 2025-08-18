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
	"github.com/reugn/go-quartz/quartz"
	"github.com/rs/zerolog"
	"github.com/samber/lo"
	"go.uber.org/atomic"

	"github.com/pubgo/lava/core/metrics"
)

type Status string

const (
	StatusInit    Status = "init"
	StatusRunning Status = "running"
	StatusStop    Status = "stop"
)

type jobWrapper struct {
	key  string
	cron string
	dur  time.Duration
	once bool
}

type namedJob struct {
	s       *Scheduler
	name    string
	fn      JobFunc
	log     log.Logger
	setting *JobSetting
	trigger *triggerImpl

	runs atomic.Uint64
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
			e.Uint64("runs", t.runs.Load())
			e.Msg("exec scheduler job")
		})
	}()

	t.runs.Inc()
	metadata := JobMetadata{
		Name:          t.setting.Name,
		Replace:       lo.FromPtr(t.setting.Replace),
		MaxRetries:    lo.FromPtr(t.setting.MaxRetries),
		RetryInterval: lo.FromPtr(t.setting.RetryInterval),
		Timeout:       lo.FromPtr(t.setting.Timeout),
		Location:      t.setting.location,
		PreRunTime:    t.trigger.prev,
		NextRunTime:   t.trigger.next,
	}

	if t.trigger.err != nil {
		return fmt.Errorf("schedule job(%s) error: %w", t.name, t.trigger.err)
	}

	return try.Try(func() error {
		ctx, cancel := context.WithTimeout(ctx, lo.FromPtr(t.setting.Timeout))
		defer cancel()

		return t.fn(ctx, t.name, &metadata)
	})
}

func registerJob(s *Scheduler, job jobWrapper, fn JobFunc) (r result.Error) {
	s.configMap[job.key] = initConfig(job.key, s.configMap[job.key]).UnwrapErr(&r)
	if r.IsErr() {
		return
	}

	setting := s.configMap[job.key]
	trigger := getTrigger(job, setting.location).UnwrapErr(&r)
	if r.IsErr() {
		return
	}

	if fn == nil {
		return result.ErrorOf("schedule job(%s) error: %s", job.key, "fn is nil")
	}

	jobOpt := &quartz.JobDetailOptions{
		MaxRetries:    lo.FromPtr(setting.MaxRetries),
		RetryInterval: lo.FromPtr(setting.RetryInterval),
		Replace:       lo.FromPtr(setting.Replace),
		Suspended:     false,
	}

	return result.ErrOf(s.scheduler.ScheduleJob(
		quartz.NewJobDetailWithOptions(
			&namedJob{s: s, name: job.key, fn: fn, log: s.log, setting: setting, trigger: trigger},
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

func getTrigger(j jobWrapper, location *time.Location) (r result.Result[*triggerImpl]) {
	if j.once {
		return r.WithValue(newTrigger(quartz.NewRunOnceTrigger(j.dur)))
	}

	if j.cron != "" {
		trigger, err := quartz.NewCronTriggerWithLoc(j.cron, location)
		if err != nil {
			return r.WithErrorf("cron-expr:%s, err:%s", j.cron, err.Error())
		}
		return r.WithValue(newTrigger(trigger))
	}

	if j.dur != 0 {
		return r.WithValue(newTrigger(quartz.NewSimpleTrigger(j.dur)))
	}

	return r.WithErrorf("please init dur or cron")
}

func parseJobKey(name string) *quartz.JobKey {
	keys := strings.SplitN(name, quartz.Sep, 2)
	if len(keys) == 1 {
		return quartz.NewJobKey(keys[0])
	}
	return quartz.NewJobKeyWithGroup(keys[0], keys[1])
}
