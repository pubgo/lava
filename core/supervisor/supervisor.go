package supervisor

import (
	"context"
	"expvar"
)

type Service interface {
	Name() string
	Error() error
	String() string
	Serve(ctx context.Context) error
	Metrics() *expvar.Map
}

type serviceFn func(ctx context.Context) error

func (fn serviceFn) Serve(ctx context.Context) error {
	return fn(ctx)
}
