package httpclient

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/pubgo/xerror"
)

// DoFunc http client do func wrapper
type DoFunc func(*http.Request, func(*http.Response) error) error

// Middleware http client middleware
type Middleware func(DoFunc) DoFunc

// Client http client interface
type Client interface {
	Get(url string, headers http.Header) (*http.Response, error)
	Post(url string, body io.Reader, headers http.Header) (*http.Response, error)
	Put(url string, body io.Reader, headers http.Header) (*http.Response, error)
	Patch(url string, body io.Reader, headers http.Header) (*http.Response, error)
	Delete(url string, headers http.Header) (*http.Response, error)
	Do(req *http.Request) (*http.Response, error)
	Options() Options
}

const (
	defaultRetryCount  = 1
	defaultHTTPTimeout = 2 * time.Second
)

var _ Client = (*client)(nil)

// client is the Client implementation
type client struct {
	client *http.Client
	opts   Options
	do     DoFunc
}

func (c *client) Options() Options {
	return c.opts
}

func newOptions(opts ...Option) Options {
	_opts := Options{
		Timeout:    defaultHTTPTimeout,
		RetryCount: defaultRetryCount,
		Retrier:    NewRetrier(NewConstantBackoff(10*time.Millisecond, 50*time.Millisecond)),
	}

	for _, opt := range opts {
		opt(&_opts)
	}

	return _opts
}

// New returns a new instance of http client
func New(opts ...Option) Client {
	cOpts := newOptions(opts...)
	c := &client{
		opts: cOpts,
		client: &http.Client{
			Timeout: cOpts.Timeout,
		},
	}

	do := c.doFunc
	for i := len(cOpts.Middles); i > 0; i-- {
		do = cOpts.Middles[i-1](do)
	}
	c.do = do

	return c
}

// Get makes a HTTP GET request to provided URL
func (c *client) Get(url string, headers http.Header) (*http.Response, error) {
	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, xerror.WrapF(err, "GET - request creation failed")
	}
	request.Header = headers

	return c.Do(request)
}

// Post makes a HTTP POST request to provided URL and requestBody
func (c *client) Post(url string, body io.Reader, headers http.Header) (*http.Response, error) {
	request, err := http.NewRequest(http.MethodPost, url, body)
	if err != nil {
		return nil, xerror.WrapF(err, "POST - request creation failed")
	}
	request.Header = headers

	return c.Do(request)
}

// Put makes a HTTP PUT request to provided URL and requestBody
func (c *client) Put(url string, body io.Reader, headers http.Header) (*http.Response, error) {
	request, err := http.NewRequest(http.MethodPut, url, body)
	if err != nil {
		return nil, xerror.WrapF(err, "PUT - request creation failed")
	}
	request.Header = headers

	return c.Do(request)
}

// Patch makes a HTTP PATCH request to provided URL and requestBody
func (c *client) Patch(url string, body io.Reader, headers http.Header) (*http.Response, error) {
	request, err := http.NewRequest(http.MethodPatch, url, body)
	if err != nil {
		return nil, xerror.WrapF(err, "PATCH - request creation failed")
	}
	request.Header = headers

	return c.Do(request)
}

// Delete makes a HTTP DELETE request with provided URL
func (c *client) Delete(url string, headers http.Header) (*http.Response, error) {
	request, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return nil, xerror.WrapF(err, "DELETE - request creation failed")
	}
	request.Header = headers

	return c.Do(request)
}

func (c *client) doFunc(req *http.Request, fn func(*http.Response) error) error {
	// nolint:bodyclose
	response, err := c.client.Do(req)
	if err != nil {
		return err
	}

	return fn(response)
}

// Do makes an HTTP request with the native `http.Do` interface
func (c *client) Do(request *http.Request) (*http.Response, error) {
	var (
		err        error
		resp       *http.Response
		bodyReader *bytes.Reader
	)

	if request.Body != nil {
		reqData, err := ioutil.ReadAll(request.Body)
		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewReader(reqData)
		request.Body = ioutil.NopCloser(bodyReader) // prevents closing the body between retries
	}

	for i := 0; i < c.opts.RetryCount; i++ {
		if resp != nil {
			resp.Body.Close()
		}

		err = c.do(request, func(response *http.Response) error {
			if bodyReader != nil {
				// Reset the body reader after the request since at this point it's already read
				// Note that it's safe to ignore the error here since the 0,0 position is always valid
				_, _ = bodyReader.Seek(0, 0)
			}
			resp = response
			return nil
		})

		if err != nil {
			if backoffTime := c.opts.Retrier.NextInterval(i); backoffTime != 0 {
				time.Sleep(backoffTime)
			}
			continue
		}

		break
	}

	return resp, err
}
