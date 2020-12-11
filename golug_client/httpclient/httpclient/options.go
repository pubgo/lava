package httpclient

import (
	"time"
)

// Options represents the http client options
type Options struct {
	Retrier    Retriable
	Timeout    time.Duration
	RetryCount int
	Middles    []Middleware
}

// Option ...
type Option func(opts *Options)
