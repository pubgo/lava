package retry

import (
	"sync/atomic"
	"time"

	"github.com/pubgo/xerror"
)

// NewExponential creates a new exponential backoff using the starting value of
// base and doubling on each failure (1, 2, 4, 8, 16, 32, 64...), up to max.
// It's very efficient, but does not check for overflow, so ensure you bound the
// retry.
func NewExponential(base time.Duration) Backoff {
	xerror.Assert(base <= 0, "base must be greater than 0")
	return &exponentialBackoff{base: base}
}

type exponentialBackoff struct {
	base    time.Duration
	attempt uint64
}

// Next implements Backoff. It is safe for concurrent use.
func (b *exponentialBackoff) Next() (time.Duration, bool) {
	return b.base << (atomic.AddUint64(&b.attempt, 1) - 1), false
}
