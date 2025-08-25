package supervisor

import (
	"context"
	"fmt"
	"time"

	"github.com/pubgo/funk/errors"
	"github.com/pubgo/funk/log"
	"github.com/pubgo/funk/recovery"
	"go.uber.org/atomic"
)

type serviceMetric struct {
	Error          atomic.String
	Restart        atomic.Uint32
	StartTime      atomic.Time
	OnlineDuration atomic.Duration
}

func NewService(name string, fn func(ctx context.Context) error) Service {
	return &serviceImpl{name: name, fn: fn, metric: &serviceMetric{}}
}

var _ Service = &serviceImpl{}

type serviceImpl struct {
	name string
	err  error
	fn   func(ctx context.Context) error

	metric *serviceMetric
}

func (s *serviceImpl) Metrics() *ServiceMetric {
	metric := s.metric
	return &ServiceMetric{
		Name:           s.name,
		Error:          metric.Error.Load(),
		Restart:        metric.Restart.Load(),
		StartTime:      metric.StartTime.Load(),
		OnlineDuration: time.Since(metric.StartTime.Load()).Truncate(time.Millisecond),
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
	s.metric.StartTime.Store(time.Now())
	defer func() {
		if gErr != nil {
			s.err = gErr
		}

		if s.err != nil {
			s.metric.Error.Store(s.err.Error())
		}

		log.Info(ctx).
			Str("service", s.name).
			Any("metrics", s.Metrics()).
			Msg("stop service")
	}()
	defer recovery.Err(&gErr)

	s.err = nil
	s.metric.Restart.Add(1)
	log.Info(ctx).Str("service", s.name).Msg("start service")
	err := s.fn(ctx)
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return fmt.Errorf("non-context error, service=%s meta=%v err=%w", s.name, s.Metrics(), err)
	}
	return err
}
