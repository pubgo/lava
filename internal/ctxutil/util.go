package ctxutil

import (
	"context"
	"time"

	"github.com/samber/lo"
)

const defaultTimeout = time.Second * 10

var _ context.Context = (*clonedCtx)(nil)

type clonedCtx struct {
	oldCtx context.Context
	ctx    context.Context
}

// Deadline implements context.Context.
func (c *clonedCtx) Deadline() (deadline time.Time, ok bool) {
	return c.ctx.Deadline()
}

// Done implements context.Context.
func (c *clonedCtx) Done() <-chan struct{} {
	return c.ctx.Done()
}

// Err implements context.Context.
func (c *clonedCtx) Err() error {
	return c.ctx.Err()
}

// Value implements context.Context.
func (c *clonedCtx) Value(key any) any {
	return c.oldCtx.Value(key)
}

func Clone(ctx context.Context, timeouts ...time.Duration) (context.Context, context.CancelFunc) {
	if ctx == nil {
		return context.WithCancel(context.Background())
	}

	timeout := defaultTimeout
	if len(timeouts) > 0 {
		timeout = timeouts[0]
	}

	deadline, ok := ctx.Deadline()
	if ok {
		timeout = time.Until(deadline)
	}

	ctxNew, cancel := context.WithTimeout(context.Background(), timeout)
	return &clonedCtx{oldCtx: ctx, ctx: ctxNew}, cancel
}

func GetTimeout(ctx context.Context) *time.Duration {
	if ctx == nil {
		return nil
	}

	deadline, ok := ctx.Deadline()
	if ok {
		return lo.ToPtr(time.Until(deadline))
	}
	return nil
}
