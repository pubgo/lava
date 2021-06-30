package retry

import (
	"time"
)

// Do wraps a function with a backoff to retry.
func Do(b Backoff, f func(i int) bool) {
	for i := 0; ; i++ {
		if f(i) {
			return
		}

		next, stop := b.Next()
		if stop {
			return
		}

		time.Sleep(next)
	}
}
