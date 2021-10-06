package ctxutil

import (
	"context"
	"time"
)

const DefaultTimeout = time.Second * 2

type Context interface {
	Context() context.Context
	Cancel() context.CancelFunc
}

var _ Context = (*ctxImpl)(nil)

type ctxImpl struct {
	context context.Context
	cancel  context.CancelFunc
}

func (c ctxImpl) Context() context.Context   { return c.context }
func (c ctxImpl) Cancel() context.CancelFunc { return c.cancel }

func Timeout(durations ...time.Duration) Context {
	var dur = DefaultTimeout
	if len(durations) > 0 {
		dur = durations[0]
	}

	var ctx, cancel = context.WithTimeout(context.Background(), dur)
	return ctxImpl{context: ctx, cancel: cancel}
}

func Default() Context {
	return ctxImpl{context: context.Background(), cancel: func() {}}
}
