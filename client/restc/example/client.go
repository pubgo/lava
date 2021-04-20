package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/pubgo/lug/client/restc"
	"github.com/pubgo/lug/client/restc/hystrix"
	"github.com/pubgo/xerror"
)

const (
	baseURL = "https://www.baidu.com"
)

func hystrixO() restc.Option {
	return restc.WithMiddleware(hystrix.Middleware(
		hystrix.WithHystrixTimeout(1100*time.Millisecond),
		hystrix.WithCommandName("MyCommand"),
		hystrix.WithMaxConcurrentRequests(100),
		hystrix.WithErrorPercentThreshold(25),
		hystrix.WithSleepWindow(10),
		hystrix.WithRequestVolumeThreshold(10),
	))
}

func httpClientUsage() error {
	timeout := 100 * time.Millisecond

	httpClient := restc.New(
		restc.WithHTTPTimeout(timeout),
		restc.WithRetryCount(2),
		restc.WithRetrier(restc.NewRetrier(restc.NewConstantBackoff(10*time.Millisecond, 50*time.Millisecond))),
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
	client := restc.New(
		restc.WithHTTPTimeout(timeout),
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

	httpclientClient := restc.New(
		restc.WithHTTPTimeout(timeout),
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
