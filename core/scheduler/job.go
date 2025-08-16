package scheduler

import (
	"context"
	"time"

	"github.com/pubgo/funk/generic"
	"github.com/pubgo/funk/log"
	"github.com/pubgo/funk/try"
	"github.com/pubgo/funk/v2/result"
	"github.com/reugn/go-quartz/quartz"
	"github.com/samber/lo"

	"github.com/pubgo/lava/core/metrics"
)

type namedJob struct {
	s       *Scheduler
	name    string
	fn      JobFunc
	log     log.Logger
	setting *JobSetting
}

func (t namedJob) Description() string { return t.name }
func (t namedJob) Execute(ctx context.Context) error {
	start := time.Now()
	metadata := JobMetadata{
		Name:          t.setting.Name,
		Replace:       lo.FromPtr(t.setting.Replace),
		MaxRetries:    lo.FromPtr(t.setting.MaxRetries),
		RetryInterval: lo.FromPtr(t.setting.RetryInterval),
		Timeout:       lo.FromPtr(t.setting.Timeout),
		Location:      t.setting.location,
	}
	err := try.Try(func() error {
		ctx, cancel := context.WithTimeout(ctx, lo.FromPtr(t.setting.Timeout))
		defer cancel()

		return t.fn(ctx, t.name, &metadata)
	})

	cost := float64(time.Since(start).Milliseconds())
	t.s.metric.Tagged(metrics.Tags{"job_name": t.name}).Gauge("job_cost_ms").Update(cost)

	logger := generic.Ternary(generic.IsNil(err), t.log.Info(), t.log.Err(err))
	logger.
		Float32("job_cost_ms", float32(cost)).
		Str("job_name", t.name).
		Msg("exec scheduler job")

	return err
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
			&namedJob{s: s, name: job.key, fn: fn, log: s.log, setting: setting},
			quartz.NewJobKey(job.key),
			jobOpt,
		),
		trigger,
	))
}

func getTrigger(j jobWrapper, location *time.Location) (r result.Result[quartz.Trigger]) {
	if j.once {
		return r.WithValue(quartz.NewRunOnceTrigger(j.dur))
	}

	if j.cron != "" {
		trigger, err := quartz.NewCronTriggerWithLoc(j.cron, location)
		if err != nil {
			return r.WithErrorf("cron-expr:%s, err:%s", j.cron, err.Error())
		}
		return r.WithValue(trigger)
	}

	if j.dur != 0 {
		return r.WithValue(quartz.NewSimpleTrigger(j.dur))
	}

	return r.WithErrorf("please init dur or cron")
}
