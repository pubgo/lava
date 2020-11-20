package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/pkg/errors"
	"github.com/pubgo/golug/golug_entry/http_entry/httpclient"
	"github.com/pubgo/golug/golug_entry/http_entry/httpclient/httphystrix"
)

const (
	baseURL = "https://www.baidu.com"
)

func hystrixO() httpclient.Option {
	return httpclient.WithMiddleware(httphystrix.Middleware(
		httphystrix.WithHystrixTimeout(1100*time.Millisecond),
		httphystrix.WithCommandName("MyCommand"),
		httphystrix.WithMaxConcurrentRequests(100),
		httphystrix.WithErrorPercentThreshold(25),
		httphystrix.WithSleepWindow(10),
		httphystrix.WithRequestVolumeThreshold(10),
	))
}

func httpClientUsage() error {
	timeout := 100 * time.Millisecond

	httpClient := httpclient.New(
		httpclient.WithHTTPTimeout(timeout),
		httpclient.WithRetryCount(2),
		httpclient.WithRetrier(httpclient.NewRetrier(httpclient.NewConstantBackoff(10*time.Millisecond, 50*time.Millisecond))),
	)

	headers := http.Header{}
	headers.Set("Content-Type", "application/json")

	response, err := httpClient.Get(baseURL, headers)
	if err != nil {
		return errors.Wrap(err, "failed to make a request to server")
	}
	defer response.Body.Close()

	respBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return errors.Wrap(err, "failed to read response body")
	}

	fmt.Printf("Response: %s", string(respBody))
	return nil
}

func httpClientClientUsage() error {
	timeout := 100 * time.Millisecond
	client := httpclient.New(
		httpclient.WithHTTPTimeout(timeout),
		hystrixO(),
	)

	headers := http.Header{}
	response, err := client.Get(baseURL, headers)
	if err != nil {
		return errors.Wrap(err, "failed to make a request to server")
	}

	defer response.Body.Close()

	respBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return errors.Wrap(err, "failed to read response body")
	}

	fmt.Printf("Response: %s", string(respBody))
	return nil
}

func customhttpclientClientUsage() error {
	timeout := 0 * time.Millisecond

	httpclientClient := httpclient.New(
		httpclient.WithHTTPTimeout(timeout),
		hystrixO(),
	)

	headers := http.Header{}
	response, err := httpclientClient.Get(baseURL, headers)
	if err != nil {
		return errors.Wrap(err, "failed to make a request to server")
	}

	defer response.Body.Close()

	respBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return errors.Wrap(err, "failed to read response body")
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
