package ctxutil

import (
	"context"
	"sync"
	"time"

	"github.com/pubgo/lava/consts"
)

var _ Ctx = (*ctxImpl)(nil)

type ctxImpl struct {
	cancel []func()
	mu     sync.RWMutex
	ctx    context.Context
}

func (c *ctxImpl) addCancel(cancel func()) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cancel = append(c.cancel, cancel)
}

func (c *ctxImpl) WithDeadline(parent context.Context, d time.Time) context.Context {
	var ctx, cancel = context.WithDeadline(parent, d)
	c.addCancel(cancel)
	return ctx
}

func (c *ctxImpl) Deadline(d time.Time) context.Context {
	var ctx, cancel = context.WithDeadline(context.Background(), d)
	c.addCancel(cancel)
	return ctx
}

func (c *ctxImpl) WithTimeout(parent context.Context, timeout time.Duration) context.Context {
	var ctx, cancel = context.WithTimeout(parent, timeout)
	c.addCancel(cancel)
	return ctx
}

func (c *ctxImpl) Timeout(timeout time.Duration) context.Context {
	var ctx, cancel = context.WithTimeout(context.Background(), timeout)
	c.addCancel(cancel)
	return ctx
}

func (c *ctxImpl) WithCancel(parent context.Context) context.Context {
	var ctx, cancel = context.WithCancel(parent)
	c.addCancel(cancel)
	return ctx
}

func (c *ctxImpl) Context() context.Context {
	var ctx, cancel = context.WithCancel(context.Background())
	c.addCancel(cancel)
	return ctx
}

func (c *ctxImpl) Cancel() {
	c.mu.RLock()
	defer c.mu.RUnlock()

	for i := range c.cancel {
		c.cancel[i]()
	}
}

func New() Ctx { return &ctxImpl{} }
func Timeout(timeout ...time.Duration) context.Context {
	if len(timeout) > 0 {
		return New().Timeout(timeout[0])
	}
	return New().Timeout(consts.DefaultTimeout)
}

func Deadline(t time.Time) context.Context { return New().Deadline(t) }
