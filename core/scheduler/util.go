package scheduler

import (
	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/errors"
	"github.com/pubgo/funk/result"
	"github.com/reugn/go-quartz/quartz"
)

func do(s *Scheduler, job job, fn JobFunc) {
	trigger := getTrigger(job)
	assert.If(job.key == "", "[name] should not be null")
	assert.If(fn == nil, "[fn] should not be nil")
	assert.If(trigger == nil, "please init dur or cron")
	assert.Must(s.scheduler.ScheduleJob(
		quartz.NewJobDetail(
			&namedJob{s: s, name: job.key, fn: fn, log: s.log},
			quartz.NewJobKey(job.key)), trigger))
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
