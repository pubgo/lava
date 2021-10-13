package ctxutil

import (
	"context"
	"time"
)

type Ctx interface {
	Cancel()
	addCancel(func())
	Context() context.Context
	WithCancel(parent context.Context) context.Context
	Deadline(d time.Time) context.Context
	WithDeadline(parent context.Context, d time.Time) context.Context
	Timeout(timeout time.Duration) context.Context
	WithTimeout(parent context.Context, timeout time.Duration) context.Context
}
