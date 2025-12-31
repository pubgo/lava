package scheduler

import (
	"strings"
	"time"

	"github.com/reugn/go-quartz/quartz"

	"github.com/pubgo/funk/v2/result"
)

func regJobExecutor(jobExecutors map[string]JobExecutor, executor JobExecutor) (r result.Error) {
	if executor == nil {
		return r.WithErrorf("executor is nil")
	}

	if executor.Name() == "" {
		return r.WithErrorf("executor name is empty")
	}

	if jobExecutors[executor.Name()] != nil {
		return r.WithErrorf("[job executor] %s already exists", executor.Name())
	}

	jobExecutors[executor.Name()] = executor
	return r
}

func getTrigger(j JobSpec, location *time.Location) (r result.Result[*triggerImpl]) {
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

	return r.WithErrorf("please init spec.Once, spec.Cron or spec.Ticker, spec:%#v", j)
}

func parseJobKey(name string) *quartz.JobKey {
	keys := strings.SplitN(name, quartz.Sep, 2)
	if len(keys) == 1 {
		return quartz.NewJobKey(keys[0])
	}
	return quartz.NewJobKeyWithGroup(keys[0], keys[1])
}
