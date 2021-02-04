package grpc_hystrix

import (
	"time"
)

type Options struct {
	HystrixTimeout         time.Duration
	HystrixCommandName     string
	MaxConcurrentRequests  int
	RequestVolumeThreshold int
	SleepWindow            int
	ErrorPercentThreshold  int
}

// Option represents the hystrix client options
type Option func(opts *Options)

// WithCommandName sets the hystrix command name
func WithCommandName(name string) Option {
	return func(c *Options) {
		c.HystrixCommandName = name
	}
}

// WithHystrixTimeout sets hystrix timeout
func WithHystrixTimeout(timeout time.Duration) Option {
	return func(c *Options) {
		c.HystrixTimeout = timeout
	}
}

// WithMaxConcurrentRequests sets hystrix max concurrent requests
func WithMaxConcurrentRequests(maxConcurrentRequests int) Option {
	return func(c *Options) {
		c.MaxConcurrentRequests = maxConcurrentRequests
	}
}

// WithRequestVolumeThreshold sets hystrix request volume threshold
func WithRequestVolumeThreshold(requestVolumeThreshold int) Option {
	return func(c *Options) {
		c.RequestVolumeThreshold = requestVolumeThreshold
	}
}

// WithSleepWindow sets hystrix sleep window
func WithSleepWindow(sleepWindow int) Option {
	return func(c *Options) {
		c.SleepWindow = sleepWindow
	}
}

// WithErrorPercentThreshold sets hystrix error percent threshold
func WithErrorPercentThreshold(errorPercentThreshold int) Option {
	return func(c *Options) {
		c.ErrorPercentThreshold = errorPercentThreshold
	}
}
