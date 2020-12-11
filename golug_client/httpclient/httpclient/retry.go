package httpclient

import (
	"time"
)

// Retriable defines contract for retriers to implement
type Retriable interface {
	NextInterval(retry int) time.Duration
}

// RetriableFunc is an adapter to allow the use of ordinary functions
// as a Retriable
type RetriableFunc func(retry int) time.Duration

// NextInterval calls f(retry)
func (f RetriableFunc) NextInterval(retry int) time.Duration {
	return f(retry)
}
