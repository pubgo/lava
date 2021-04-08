package rest

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

// WithHTTPTimeout sets hystrix timeout
func WithHTTPTimeout(timeout time.Duration) Option {
	return func(c *Options) {
		c.Timeout = timeout
	}
}

// WithRetryCount sets the retry count for the client
func WithRetryCount(retryCount int) Option {
	return func(c *Options) {
		c.RetryCount = retryCount
	}
}

// WithRetrier sets the strategy for retrying
func WithRetrier(retrier Retriable) Option {
	return func(c *Options) {
		c.Retrier = retrier
	}
}

// WithMiddleware sets the strategy for retrying
func WithMiddleware(m Middleware) Option {
	return func(c *Options) {
		c.Middles = append(c.Middles, m)
	}
}
