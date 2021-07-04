package hystrix

import (
	restc2 "github.com/pubgo/lug/plugins/restc"
	"time"

	"github.com/afex/hystrix-go/hystrix"
)

const (
	defaultHystrixTimeout         = 2 * time.Second
	defaultMaxConcurrentRequests  = 5000
	defaultErrorPercentThreshold  = 25
	defaultSleepWindow            = 10
	defaultRequestVolumeThreshold = 10
	defaultCommandName            = "http.client"
	maxUint                       = ^uint(0)
	maxInt                        = int(maxUint >> 1)
)

func Middleware(opts ...Option) restc2.Middleware {
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

	return func(doFunc restc2.DoFunc) restc2.DoFunc {
		return func(req *restc2.Request, fn func(resp *restc2.Response) error) error {
			return hystrix.Do(hOpts.HystrixCommandName, func() error {
				return doFunc(req, fn)
			}, nil)
		}
	}
}

func durationToInt(duration, unit time.Duration) int {
	durationAsNumber := duration / unit

	if int64(durationAsNumber) > int64(maxInt) {
		// Returning max possible value seems like best possible solution here
		// the alternative is to panic as there is no way of returning an error
		// without changing the NewClient API
		return maxInt
	}
	return int(durationAsNumber)
}