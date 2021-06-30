package restc

import (
	"net/url"
	"time"

	"github.com/pubgo/lug/pkg/retry"
	"github.com/pubgo/xerror"
	"github.com/valyala/fasthttp"
)

const (
	defaultRetryCount  = 1
	defaultHTTPTimeout = 2 * time.Second
	defaultContentType = "application/json"
)

var _ Client = (*client)(nil)

// client is the Client implementation
type client struct {
	client        *fasthttp.Client
	cfg           Cfg
	do            DoFunc
	defaultHeader *fasthttp.RequestHeader
}

func (c *client) Do(req *Request) (resp *Response, err error) {
	return resp, xerror.Wrap(c.do(req, func(res *Response) error {
		resp = res
		return nil
	}))
}

func doUrl(c *client, mth string, url string, requests ...func(req *Request)) (*Response, error) {
	var req = fasthttp.AcquireRequest()
	c.defaultHeader.CopyTo(&req.Header)

	req.SetRequestURI(url)
	req.Header.SetMethod(mth)
	req.Header.SetContentType(defaultContentType)

	if len(requests) > 0 {
		requests[0](req)
	}

	var resp, err = c.Do(req)
	fasthttp.ReleaseRequest(req)

	if err != nil {
		return nil, xerror.Wrap(err)
	}

	return resp, nil
}

func (c *client) Get(url string, requests ...func(req *Request)) (*Response, error) {
	var resp, err = doUrl(c, fasthttp.MethodGet, url, requests...)
	return resp, xerror.Wrap(err)
}

func (c *client) Delete(url string, requests ...func(req *Request)) (*Response, error) {
	var resp, err = doUrl(c, fasthttp.MethodDelete, url, requests...)
	return resp, xerror.Wrap(err)
}

func (c *client) Post(url string, requests ...func(req *Request)) (*Response, error) {
	var resp, err = doUrl(c, fasthttp.MethodDelete, url, requests...)
	return resp, xerror.Wrap(err)
}

func (c *client) PostForm(url string, val url.Values, requests ...func(req *Request)) (*Response, error) {
	var resp, err = doUrl(c, fasthttp.MethodDelete, url, func(req *Request) {
		req.SetBodyString(val.Encode())
		req.Header.SetContentType("application/x-www-form-urlencoded")
		if len(requests) > 0 {
			requests[0](req)
		}
	})
	return resp, xerror.Wrap(err)
}

func (c *client) Put(url string, requests ...func(req *Request)) (*Response, error) {
	var resp, err = doUrl(c, fasthttp.MethodPut, url, requests...)
	return resp, xerror.Wrap(err)
}

func (c *client) Patch(url string, requests ...func(req *Request)) (*Response, error) {
	var resp, err = doUrl(c, fasthttp.MethodPatch, url, requests...)
	return resp, xerror.Wrap(err)
}

func doFunc(c *client) func(req *Request, fn func(*Response) error) error {
	return func(req *Request, fn func(*Response) error) (err error) {
		var resp = fasthttp.AcquireResponse()
		retry.Do(retry.WithMaxRetries(c.cfg.RetryCount, c.cfg.backoff), func(i int) bool {
			if c.cfg.Timeout > 0 {
				err = c.client.DoTimeout(req, resp, c.cfg.Timeout)
			} else {
				err = c.client.Do(req, resp)
			}

			if err != nil {
				return false
			}

			return true
		})

		if err != nil {
			return xerror.Wrap(err)
		}

		return xerror.Wrap(fn(resp))
	}
}
