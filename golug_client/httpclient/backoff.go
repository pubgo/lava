package httpclient

import (
	"time"
)

// Backoff interface defines contract for backoff strategies
type Backoff interface {
	Next(retry int) time.Duration
}
