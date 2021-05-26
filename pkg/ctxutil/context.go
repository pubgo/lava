package ctxutil

import (
	"context"
	"time"
)

const DefaultTimeout = time.Second * 2

func Timeout(durations ...time.Duration) (context.Context, context.CancelFunc) {
	var dur = DefaultTimeout
	if len(durations) > 0 {
		dur = durations[0]
	}

	return context.WithTimeout(context.Background(), dur)
}

func Default() context.Context {
	return context.Background()
}
