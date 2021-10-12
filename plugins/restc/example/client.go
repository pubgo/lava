package main

import (
	"context"
	"fmt"
	"time"

	"github.com/pubgo/xerror"
	"github.com/valyala/fasthttp"

	"github.com/pubgo/lava/pkg/retry"
	"github.com/pubgo/lava/plugins/hystrix"
	"github.com/pubgo/lava/plugins/restc"
)

const (
	baseURL = "https://www.cnblogs.com/bergus/articles/nginx-kai-qi-response-heheader-ri-zhi-ji-lu.html"
)

func hystrixO() restc.Option {
	return restc.WithMiddle(hystrix.Middleware(
		hystrix.WithHystrixTimeout(1100*time.Millisecond),
		hystrix.WithCommandName("MyCommand"),
		hystrix.WithMaxConcurrentRequests(100),
		hystrix.WithErrorPercentThreshold(25),
		hystrix.WithSleepWindow(10),
		hystrix.WithRequestVolumeThreshold(10),
	))
}

func httpClientUsage() error {
	var cfg = restc.DefaultCfg()
	httpClient, err := cfg.Build(
		hystrixO(),
		restc.WithBackoff(retry.NewConstant(10*time.Millisecond)),
	)
	xerror.Panic(err)

	response, err := httpClient.Get(context.Background(), baseURL)
	defer fasthttp.ReleaseResponse(response)
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
