package scheduler

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/pubgo/funk/v2"
	"github.com/pubgo/funk/v2/clone"
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

	t.s.mu.RLock()
	name := t.task.spec.Name
	config := clone.Clone(t.task.spec.Config)
	t.s.mu.RUnlock()

	if config == nil {
		return fmt.Errorf("schedule job(%s) config is nil", name)
	}

	preExecTime, nextExecTime, triggerErr := t.task.trigger.Snapshot()

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

	location := time.UTC
	if config.location != nil {
		location = config.location
	}

	metadata := JobMetadata{
		Name:          config.Name,
		Replace:       lo.FromPtr(config.Replace),
		MaxRetries:    lo.FromPtr(config.MaxRetries),
		RetryInterval: lo.FromPtr(config.RetryInterval),
		Timeout:       lo.FromPtr(config.Timeout),
		Location:      location.String(),
		ExecTime:      preExecTime,
		NextExecTime:  nextExecTime,
	}

	// 检查 trigger 错误，但对于一次性任务的 ErrTriggerExpired 忽略
	if triggerErr != nil {
		isOnceJobExpired := errors.Is(triggerErr, quartz.ErrTriggerExpired) && t.task.spec.Once != nil
		if !isOnceJobExpired {
			return fmt.Errorf("schedule job(%s) trigger error: %w", name, triggerErr)
		}
	}

	return try.Try(func() error {
		ctx, cancel := context.WithTimeout(ctx, lo.FromPtr(config.Timeout))
		defer cancel()

		res := t.task.executor.Exec(ctx, name, &metadata)
		t.task.setResult(res)
		return res.GetErr()
	})
}

var _ quartz.Trigger = &triggerImpl{}

func newTrigger(trigger quartz.Trigger) *triggerImpl {
	return &triggerImpl{trigger: trigger}
}

type triggerImpl struct {
	mu      sync.RWMutex
	prev    int64
	next    int64
	err     error
	trigger quartz.Trigger
}

func (t *triggerImpl) NextFireTime(prev int64) (next int64, err error) {
	next, err = t.trigger.NextFireTime(prev)

	t.mu.Lock()
	t.err = err
	if err == nil {
		// 保留毫秒精度
		t.prev = prev / 1_000_000
		t.next = next / 1_000_000
	}
	t.mu.Unlock()

	return next, err
}

func (t *triggerImpl) Description() string {
	return t.trigger.Description()
}

func (t *triggerImpl) Snapshot() (prev int64, next int64, err error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	return t.prev, t.next, t.err
}
