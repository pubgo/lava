package main

import (
	"fmt"
	restc2 "github.com/pubgo/lug/restc"
	hystrix2 "github.com/pubgo/lug/restc/hystrix"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/pubgo/xerror"
)

const (
	baseURL = "https://www.baidu.com"
)

func hystrixO() restc2.Option {
	return restc2.WithMiddleware(hystrix2.Middleware(
		hystrix2.WithHystrixTimeout(1100*time.Millisecond),
		hystrix2.WithCommandName("MyCommand"),
		hystrix2.WithMaxConcurrentRequests(100),
		hystrix2.WithErrorPercentThreshold(25),
		hystrix2.WithSleepWindow(10),
		hystrix2.WithRequestVolumeThreshold(10),
	))
}

func httpClientUsage() error {
	timeout := 100 * time.Millisecond

	httpClient := restc2.New(
		restc2.WithHTTPTimeout(timeout),
		restc2.WithRetryCount(2),
		restc2.WithRetrier(restc2.NewRetrier(restc2.NewConstantBackoff(10*time.Millisecond, 50*time.Millisecond))),
	)

	headers := http.Header{}
	headers.Set("Content-Type", "application/json")

	response, err := httpClient.Get(baseURL, headers)
	if err != nil {
		return xerror.Wrap(err, "failed to make a request to server")
	}
	defer response.Body.Close()

	respBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return xerror.Wrap(err, "failed to read response body")
	}

	fmt.Printf("Response: %s", string(respBody))
	return nil
}

func httpClientClientUsage() error {
	timeout := 100 * time.Millisecond
	client := restc2.New(
		restc2.WithHTTPTimeout(timeout),
		hystrixO(),
	)

	headers := http.Header{}
	response, err := client.Get(baseURL, headers)
	if err != nil {
		return xerror.Wrap(err, "failed to make a request to server")
	}

	defer response.Body.Close()

	respBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return xerror.Wrap(err, "failed to read response body")
	}

	fmt.Printf("Response: %s", string(respBody))
	return nil
}

func customhttpclientClientUsage() error {
	timeout := 0 * time.Millisecond

	httpclientClient := restc2.New(
		restc2.WithHTTPTimeout(timeout),
		hystrixO(),
	)

	headers := http.Header{}
	response, err := httpclientClient.Get(baseURL, headers)
	if err != nil {
		return xerror.Wrap(err, "failed to make a request to server")
	}

	defer response.Body.Close()

	respBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return xerror.Wrap(err, "failed to read response body")
	}

	fmt.Printf("Response: %s", string(respBody))
	return nil
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	check(httpClientUsage())
	check(httpClientClientUsage())
	check(customhttpclientClientUsage())
}
