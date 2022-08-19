package retry

import (
	"time"
)

// NewNoop creates a new noop backoff
func NewNoop() Backoff {
	return BackoffFunc(func() (time.Duration, bool) { return 0, false })
}
