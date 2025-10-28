package scheduler

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/pubgo/funk/v2"
	"github.com/pubgo/funk/v2/log"
	"github.com/pubgo/funk/v2/try"
	"github.com/reugn/go-quartz/quartz"
	"github.com/rs/zerolog"
	"github.com/samber/lo"

	"github.com/pubgo/lava/v2/core/metrics"
)

type namedJob struct {
	s   *Scheduler
	log log.Logger

	task *jobTask
}

func (t *namedJob) Description() string { return t.task.spec.Name }
func (t *namedJob) Execute(ctx context.Context) (gErr error) {
	start := time.Now()
	name := t.task.spec.Name

	defer func() {
		cost := float64(time.Since(start).Milliseconds())
		t.s.metric.Tagged(metrics.Tags{"job_name": name}).Gauge("job_cost_ms").Update(cost)

		logger := funk.Ternary(gErr == nil, t.log.Info(), t.log.Err(gErr))
		logger.Func(func(e *zerolog.Event) {
			e.Float32("job_cost_ms", float32(cost))
			e.Str("job_name", name)
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
		Location:      config.location.String(),
		ExecTime:      t.task.trigger.prev,
		NextExecTime:  t.task.trigger.next,
	}

	if t.task.trigger.err != nil {
		if !errors.Is(t.task.trigger.err, quartz.ErrTriggerExpired) || t.task.spec.Once == nil {
			return fmt.Errorf("schedule job(%s) trigger error: %w", name, t.task.trigger.err)
		}
	}

	return try.Try(func() error {
		ctx, cancel := context.WithTimeout(ctx, lo.FromPtr(config.Timeout))
		defer cancel()

		t.task.result = t.task.executor.Exec(ctx, name, &metadata)
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
	defer func() {
		t.err = err
		if err != nil {
			return
		}

		t.prev = prev / 1000_000_000
		t.next = next / 1000_000_000
	}()

	return t.trigger.NextFireTime(prev)
}

func (t *triggerImpl) Description() string {
	return t.trigger.Description()
}
