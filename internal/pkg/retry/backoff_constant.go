package retry

import (
	"time"

	"github.com/pubgo/xerror"
)

const DefaultConstant = time.Second

// NewConstant creates a new constant backoff using the value t.
func NewConstant(t time.Duration) Backoff {
	xerror.Assert(t <= 0, "[t] must be greater than 0")

	return BackoffFunc(func() (time.Duration, bool) { return t, false })
}
