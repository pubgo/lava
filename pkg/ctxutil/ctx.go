package ctxutil

import (
	"context"
	"time"

	"github.com/pubgo/lava/consts"
)

func Timeout(timeout ...time.Duration) context.Context {
	t := consts.DefaultTimeout
	if len(timeout) > 0 {
		t = timeout[0]
	}
	var ctx, cancel = context.WithTimeout(context.Background(), t)
	_ = cancel
	return ctx
}
