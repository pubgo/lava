package restc

import (
	"context"
	"github.com/pubgo/lava/plugins/tracing"
	"net/url"
	"time"

	"github.com/opentracing/opentracing-go/ext"
	"github.com/pubgo/xerror"
	"github.com/valyala/fasthttp"

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
	client        *fasthttp.Client
	defaultHeader *fasthttp.RequestHeader
	cfg           Cfg
	do            types.MiddleNext
}

func (c *clientImpl) Do(ctx context.Context, req *Request) (resp *Response, err error) {
	return resp, xerror.Wrap(c.do(ctx,
		&request{req: req, header: convertHeader(&req.Header)},
		func(res types.Response) error {
			resp = res.(*response).resp
			return nil
		},
	))
}

func (c *clientImpl) Get(ctx context.Context, url string, requests ...func(req *Request)) (*Response, error) {
	return doUrl(ctx, c, fasthttp.MethodGet, url, requests...)
}

func (c *clientImpl) Delete(ctx context.Context, url string, requests ...func(req *Request)) (*Response, error) {
	return doUrl(ctx, c, fasthttp.MethodDelete, url, requests...)
}

func (c *clientImpl) Post(ctx context.Context, url string, requests ...func(req *Request)) (*Response, error) {
	return doUrl(ctx, c, fasthttp.MethodPost, url, requests...)
}

func (c *clientImpl) PostForm(ctx context.Context, url string, val url.Values, requests ...func(req *Request)) (*Response, error) {
	var resp, err = doUrl(ctx, c, fasthttp.MethodPost, url, func(req *Request) {
		req.SetBodyString(val.Encode())
		req.Header.SetContentType("application/x-www-form-urlencoded")
		if len(requests) > 0 {
			requests[0](req)
		}
	})
	return resp, xerror.Wrap(err)
}

func (c *clientImpl) Put(ctx context.Context, url string, requests ...func(req *Request)) (*Response, error) {
	return doUrl(ctx, c, fasthttp.MethodPut, url, requests...)
}

func (c *clientImpl) Patch(ctx context.Context, url string, requests ...func(req *Request)) (*Response, error) {
	return doUrl(ctx, c, fasthttp.MethodPatch, url, requests...)
}

func doUrl(ctx context.Context, c *clientImpl, mth string, url string, requests ...func(req *Request)) (*Response, error) {
	var req = fasthttp.AcquireRequest()
	c.defaultHeader.CopyTo(&req.Header)

	req.SetRequestURI(url)
	req.Header.SetMethod(mth)

	if len(requests) > 0 {
		requests[0](req)
	}

	var resp, err = c.Do(ctx, req)
	fasthttp.ReleaseRequest(req)

	if err != nil {
		return nil, xerror.WrapF(err, "method=>%s, url=>%s", mth, url)
	}

	return resp, nil
}

func doFunc(c *clientImpl) types.MiddleNext {
	var r = retry.New(retry.WithMaxRetries(c.cfg.RetryCount, c.cfg.backoff))
	return func(ctx context.Context, req types.Request, callback func(rsp types.Response) error) error {
		var resp = fasthttp.AcquireResponse()

		defer func() {
			var span = tracing.FromCtx(ctx)
			ext.HTTPStatusCode.Set(span, uint16(resp.StatusCode()))
		}()

		xerror.Panic(r.Do(func(i int) error {
			if c.cfg.Timeout > 0 {
				return xerror.Wrap(c.client.DoTimeout(req.(*request).req, resp, c.cfg.Timeout))
			}

			return xerror.Wrap(c.client.Do(req.(*request).req, resp))
		}))

		return xerror.Wrap(callback(&response{resp: resp}))
	}
}
