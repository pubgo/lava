package grpc_hystrix

import (
	"context"
	"path"
	"time"

	"github.com/afex/hystrix-go/hystrix"
	"google.golang.org/grpc"
)

const (
	defaultHystrixTimeout         = 2 * time.Second
	defaultMaxConcurrentRequests  = 5000
	defaultErrorPercentThreshold  = 25
	defaultSleepWindow            = 10
	defaultRequestVolumeThreshold = 10
	defaultCommandName            = "grpc.client"
	maxUint                       = ^uint(0)
	maxInt                        = int(maxUint >> 1)
)

func UnaryClientInterceptor(opts ...Option) grpc.UnaryClientInterceptor {
	hOpts := Options{
		HystrixCommandName:     defaultCommandName,
		HystrixTimeout:         defaultHystrixTimeout,
		MaxConcurrentRequests:  defaultMaxConcurrentRequests,
		ErrorPercentThreshold:  defaultErrorPercentThreshold,
		SleepWindow:            defaultSleepWindow,
		RequestVolumeThreshold: defaultRequestVolumeThreshold,
	}

	for _, opt := range opts {
		opt(&hOpts)
	}

	hystrix.ConfigureCommand(
		hOpts.HystrixCommandName,
		hystrix.CommandConfig{
			Timeout:                durationToInt(hOpts.HystrixTimeout, time.Millisecond),
			MaxConcurrentRequests:  hOpts.MaxConcurrentRequests,
			RequestVolumeThreshold: hOpts.RequestVolumeThreshold,
			SleepWindow:            hOpts.SleepWindow,
			ErrorPercentThreshold:  hOpts.ErrorPercentThreshold,
		},
	)

	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		breakerName := path.Join(cc.Target(), method)
		return hystrix.Do(breakerName, func() error { return invoker(ctx, method, req, reply, cc, opts...) }, nil)
	}
}

func StreamClientInterceptor(opts ...Option) grpc.StreamClientInterceptor {
	hOpts := Options{
		HystrixCommandName:     defaultCommandName,
		HystrixTimeout:         defaultHystrixTimeout,
		MaxConcurrentRequests:  defaultMaxConcurrentRequests,
		ErrorPercentThreshold:  defaultErrorPercentThreshold,
		SleepWindow:            defaultSleepWindow,
		RequestVolumeThreshold: defaultRequestVolumeThreshold,
	}

	for _, opt := range opts {
		opt(&hOpts)
	}

	hystrix.ConfigureCommand(
		hOpts.HystrixCommandName,
		hystrix.CommandConfig{
			Timeout:                durationToInt(hOpts.HystrixTimeout, time.Millisecond),
			MaxConcurrentRequests:  hOpts.MaxConcurrentRequests,
			RequestVolumeThreshold: hOpts.RequestVolumeThreshold,
			SleepWindow:            hOpts.SleepWindow,
			ErrorPercentThreshold:  hOpts.ErrorPercentThreshold,
		},
	)

	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (cs grpc.ClientStream, err error) {
		breakerName := path.Join(cc.Target(), method)
		return cs, hystrix.Do(breakerName, func() error { cs, err = streamer(ctx, desc, cc, method, opts...); return err }, nil)
	}
}

func durationToInt(duration, unit time.Duration) int {
	durationAsNumber := duration / unit

	if int64(durationAsNumber) > int64(maxInt) {
		// Returning max possible value seems like best possible solution here
		// the alternative is to panic as there is no way of returning an error
		// without changing the GetClient API
		return maxInt
	}
	return int(durationAsNumber)
}
