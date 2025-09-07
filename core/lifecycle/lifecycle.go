package lifecycle

import "context"

func WrapNoError(fn func(context.Context)) ExecFunc {
	return func(ctx context.Context) error { fn(ctx); return nil }
}

func WrapNoCtx(fn func() error) ExecFunc {
	return func(ctx context.Context) error { return fn() }
}

func WrapNoCtxErr(fn func()) ExecFunc {
	return func(ctx context.Context) error { fn(); return nil }
}

type ExecFunc = func(context.Context) error

type Executor struct {
	Exec ExecFunc
}

type Handler func(lc Lifecycle)

type Lifecycle interface {
	AfterStop(f ExecFunc)
	BeforeStop(f ExecFunc)
	AfterStart(f ExecFunc)
	BeforeStart(f ExecFunc)
}

type Getter interface {
	GetAfterStops() []Executor
	GetBeforeStops() []Executor
	GetAfterStarts() []Executor
	GetBeforeStarts() []Executor
}
