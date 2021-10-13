package async

import (
	"context"
	"sync"
	"time"
)

type Ctx interface {
	Cancel()
	addCancel(func())
	Context() context.Context
	WithCancel(parent context.Context) context.Context
	WithDeadline(parent context.Context, d time.Time) context.Context
	WithTimeout(parent context.Context, timeout time.Duration) context.Context
}

var _ Ctx = (*ctxImpl)(nil)

type ctxImpl struct {
	cancel []func()
	mu     sync.RWMutex
}

func (c *ctxImpl) Context() context.Context {
	var ctx, cancel = context.WithCancel(context.Background())
	c.addCancel(cancel)
	return ctx
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

func (c *ctxImpl) WithTimeout(parent context.Context, timeout time.Duration) context.Context {
	var ctx, cancel = context.WithTimeout(parent, timeout)
	c.addCancel(cancel)
	return ctx
}

func (c *ctxImpl) WithCancel(parent context.Context) context.Context {
	var ctx, cancel = context.WithCancel(parent)
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
