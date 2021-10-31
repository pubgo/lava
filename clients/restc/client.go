package restc

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/pubgo/x/strutil"
	"github.com/pubgo/xerror"

	"github.com/pubgo/lava/pkg/httpx"
	"github.com/pubgo/lava/pkg/retry"
	"github.com/pubgo/lava/types"
)

const (
	defaultRetryCount  = 1
	defaultHTTPTimeout = 2 * time.Second
	defaultContentType = "application/json"
)

var _ Client = (*clientImpl)(nil)

// clientImpl is the Client implementation
type clientImpl struct {
	client *http.Client
	cfg    Cfg
	do     types.MiddleNext
}

func (c *clientImpl) Do(ctx context.Context, req *Request) (resp *Response, err error) {
	return resp, c.do(ctx, &request{req: req}, func(res types.Response) error {
		resp = res.(*response).resp
		return nil
	})
}

func (c *clientImpl) Get(ctx context.Context, url string, requests ...func(req *Request)) (*Response, error) {
	return doRequest(ctx, c, http.MethodGet, url, requests...)
}

func (c *clientImpl) Delete(ctx context.Context, url string, requests ...func(req *Request)) (*Response, error) {
	return doRequest(ctx, c, http.MethodDelete, url, requests...)
}

func (c *clientImpl) Post(ctx context.Context, url string, requests ...func(req *Request)) (*Response, error) {
	return doRequest(ctx, c, http.MethodPost, url, requests...)
}

func (c *clientImpl) PostForm(ctx context.Context, url string, val url.Values, requests ...func(req *Request)) (*Response, error) {
	var resp, err = doRequest(ctx, c, http.MethodPost, url, func(req *Request) {
		req.Header.Set(httpx.HeaderContentType, "application/x-www-form-urlencoded")
		req.GetBody = func() (io.ReadCloser, error) {
			return io.NopCloser(bytes.NewReader(strutil.ToBytes(val.Encode()))), nil
		}

		if len(requests) > 0 {
			requests[0](req)
		}
	})
	return resp, err
}

func (c *clientImpl) Put(ctx context.Context, url string, requests ...func(req *Request)) (*Response, error) {
	return doRequest(ctx, c, http.MethodPut, url, requests...)
}

func (c *clientImpl) Patch(ctx context.Context, url string, requests ...func(req *Request)) (*Response, error) {
	return doRequest(ctx, c, http.MethodPatch, url, requests...)
}

func doRequest(ctx context.Context, c *clientImpl, mth string, url string, requests ...func(req *Request)) (*Response, error) {
	req, err := http.NewRequestWithContext(ctx, mth, url, nil)
	if err != nil {
		return nil, xerror.Wrap(err, mth, url)
	}

	req.Header.Set(httpx.HeaderContentType, defaultContentType)

	if len(requests) > 0 {
		requests[0](req)
	}

	resp, err := c.Do(ctx, req)
	if err != nil {
		return nil, xerror.Wrap(err, mth, url)
	}

	return resp, nil
}

func doFunc(c *clientImpl) types.MiddleNext {
	var r = retry.New(retry.WithMaxRetries(c.cfg.RetryCount, c.cfg.backoff))
	return func(ctx context.Context, req types.Request, callback func(rsp types.Response) error) error {
		return r.Do(func(i int) error {
			resp, err := c.client.Do(req.(*request).req)
			if err != nil {
				return err
			}
			return callback(&response{resp: resp})
		})
	}
}
