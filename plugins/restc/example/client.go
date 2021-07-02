package main

import (
	"fmt"
	restc2 "github.com/pubgo/lug/plugins/restc"
	hystrix2 "github.com/pubgo/lug/plugins/restc/hystrix"
	"time"

	"github.com/pubgo/lug/pkg/retry"
	"github.com/pubgo/xerror"
)

const (
	baseURL = "https://www.cnblogs.com/bergus/articles/nginx-kai-qi-response-heheader-ri-zhi-ji-lu.html"
)

func hystrixO() restc2.Option {
	return restc2.WithMiddle(hystrix2.Middleware(
		hystrix2.WithHystrixTimeout(1100*time.Millisecond),
		hystrix2.WithCommandName("MyCommand"),
		hystrix2.WithMaxConcurrentRequests(100),
		hystrix2.WithErrorPercentThreshold(25),
		hystrix2.WithSleepWindow(10),
		hystrix2.WithRequestVolumeThreshold(10),
	))
}

func httpClientUsage() error {
	var cfg = restc2.DefaultCfg()
	httpClient, err := cfg.Build(
		hystrixO(),
		restc2.WithBackoff(retry.NewConstant(10*time.Millisecond)),
	)
	xerror.Panic(err)

	response, err := httpClient.Get(baseURL)
	defer restc2.ReleaseResponse(response)
	if err != nil {
		return xerror.Wrap(err, "failed to make a request to server")
	}

	fmt.Printf("Response: %s", string(response.Body()))
	return nil
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	check(httpClientUsage())
}
