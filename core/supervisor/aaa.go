package supervisor

import (
	"context"
	"time"

	"github.com/thejerf/suture/v4"
)

type Supervisor = suture.Supervisor

type ServiceMetric struct {
	Name           string
	Error          string
	Restart        uint32
	StartTime      time.Time
	OnlineDuration time.Duration
}

type Service interface {
	Name() string
	Error() error
	String() string
	Serve(ctx context.Context) error
	Metrics() *ServiceMetric
}

type serviceFn func(ctx context.Context) error

func (fn serviceFn) Serve(ctx context.Context) error {
	return fn(ctx)
}
