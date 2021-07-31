package retry

import (
	"github.com/pubgo/xerror"

	"time"
)

const DefaultConstant = time.Second

// NewConstant creates a new constant backoff using the value t. The wait time
// is the provided constant value.
func NewConstant(t time.Duration) Backoff {
	xerror.Assert(t <= 0, "[t] must be greater than 0")

	return BackoffFunc(func() (time.Duration, bool) { return t, false })
}
