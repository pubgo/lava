package scheduler

import (
	"context"
	"fmt"
	"time"

	"github.com/pubgo/funk/generic"
	"github.com/pubgo/funk/log"
	"github.com/pubgo/funk/try"
	"github.com/pubgo/lava/core/metrics"
	"github.com/reugn/go-quartz/quartz"
	"github.com/rs/zerolog"
	"github.com/samber/lo"
)

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
	config := t.task.spec.Config
	metadata := JobMetadata{
		Name:          config.Name,
		Replace:       lo.FromPtr(config.Replace),
		MaxRetries:    lo.FromPtr(config.MaxRetries),
		RetryInterval: lo.FromPtr(config.RetryInterval),
		Timeout:       lo.FromPtr(config.Timeout),
		Location:      config.location,
		PreRunTime:    t.task.trigger.prev,
		ExecTime:      t.task.trigger.next,
	}

	if t.task.trigger.err != nil {
		return fmt.Errorf("schedule job(%s) error: %w", t.name, t.task.trigger.err)
	}

	return try.Try(func() error {
		ctx, cancel := context.WithTimeout(ctx, lo.FromPtr(config.Timeout))
		defer cancel()

		t.task.result = t.task.executor.Exec(ctx, t.name, &metadata)
		return t.task.result.GetErr()
	})
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
