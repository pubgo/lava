package supervisor

import (
	"context"
	"fmt"
	"time"

	"github.com/pubgo/funk/v2/errors"
	"github.com/pubgo/funk/v2/log"
	"github.com/pubgo/funk/v2/recovery"
	"go.uber.org/atomic"
)

type serviceMetric struct {
	Status        atomic.String   // 当前状态
	StartCount    atomic.Uint32   // 启动次数
	ErrorCount    atomic.Uint32   // 错误次数
	SuccessCount  atomic.Uint32   // 成功退出次数
	LastError     atomic.String   // 最后一次错误信息
	LastErrorTime atomic.Time     // 最后一次错误时间
	LastStartTime atomic.Time     // 最后一次启动时间
	LastStopTime  atomic.Time     // 最后一次停止时间
	TotalUptime   atomic.Duration // 总运行时长
	CreatedAt     atomic.Time     // 服务创建时间
}

func NewService(name string, fn func(ctx context.Context) error) Service {
	m := &serviceMetric{}
	m.Status.Store(string(StatusIdle))
	m.CreatedAt.Store(time.Now())
	return &serviceImpl{name: name, fn: fn, metric: m}
}

var _ Service = &serviceImpl{}

type serviceImpl struct {
	name string
	err  error
	fn   func(ctx context.Context) error

	metric *serviceMetric
}

func (s *serviceImpl) Metric() *Metric {
	m := s.metric
	status := ServiceStatus(m.Status.Load())
	startCount := m.StartCount.Load()
	lastStartTime := m.LastStartTime.Load()
	totalUptime := m.TotalUptime.Load()

	// 计算当前运行时长
	var currentUptime time.Duration
	if status == StatusRunning && !lastStartTime.IsZero() {
		currentUptime = time.Since(lastStartTime)
	}

	// 计算平均运行时长
	var averageUptime time.Duration
	if startCount > 0 {
		averageUptime = (totalUptime + currentUptime) / time.Duration(startCount)
	}

	return &Metric{
		Name:          s.name,
		Status:        status,
		StartCount:    startCount,
		ErrorCount:    m.ErrorCount.Load(),
		SuccessCount:  m.SuccessCount.Load(),
		LastError:     m.LastError.Load(),
		LastErrorTime: m.LastErrorTime.Load(),
		LastStartTime: lastStartTime,
		LastStopTime:  m.LastStopTime.Load(),
		CurrentUptime: currentUptime,
		TotalUptime:   totalUptime + currentUptime,
		AverageUptime: averageUptime,
		CreatedAt:     m.CreatedAt.Load(),
	}
}

func (s *serviceImpl) Error() error {
	return s.err
}

func (s *serviceImpl) Name() string {
	return s.name
}

func (s *serviceImpl) String() string {
	return s.name
}

func (s *serviceImpl) Serve(ctx context.Context) (gErr error) {
	startTime := time.Now()
	s.metric.LastStartTime.Store(startTime)
	s.metric.Status.Store(string(StatusRunning))
	s.metric.StartCount.Add(1)

	defer func() {
		stopTime := time.Now()
		runDuration := stopTime.Sub(startTime)

		// 累加总运行时长
		s.metric.TotalUptime.Add(runDuration)
		s.metric.LastStopTime.Store(stopTime)

		// context 取消或超时视为正常停止
		isNormalStop := gErr == nil ||
			errors.Is(gErr, context.Canceled) ||
			errors.Is(gErr, context.DeadlineExceeded)

		if isNormalStop {
			s.metric.Status.Store(string(StatusStopped))
			s.metric.SuccessCount.Add(1)
		} else {
			s.err = gErr
			s.metric.Status.Store(string(StatusError))
			s.metric.ErrorCount.Add(1)
			s.metric.LastError.Store(gErr.Error())
			s.metric.LastErrorTime.Store(stopTime)
		}

		log.Info(ctx).
			Str("service", s.name).
			Any("metrics", s.Metric()).
			Msg("stop service")
	}()
	defer recovery.Err(&gErr)

	s.err = nil
	log.Info(ctx).Str("service", s.name).Msg("start service")
	err := s.fn(ctx)
	if err != nil && !errors.Is(err, context.Canceled) && !errors.Is(err, context.DeadlineExceeded) {
		return fmt.Errorf("service error, service=%s meta=%v err=%w", s.name, s.Metric(), err)
	}
	return err
}
