package restc

import (
	"context"
	"github.com/pubgo/lava/core/middleware"
	"net/http"
	"net/url"
	"time"

	"github.com/pubgo/x/strutil"
	"github.com/pubgo/xerror"
	"github.com/valyala/fasthttp"

	"github.com/pubgo/lava/pkg/httpx"
	"github.com/pubgo/lava/pkg/utils"
	"github.com/pubgo/lava/runtime"
)

const (
	defaultRetryCount  = 1
	defaultHTTPTimeout = 2 * time.Second
	defaultContentType = "application/json"
)

var _ Client = (*clientImpl)(nil)

// clientImpl is the Client implementation
type clientImpl struct {
	client  *fasthttp.Client
	cfg     Cfg
	do      middleware.HandlerFunc
	plugins []middleware.Middleware
}

func (c *clientImpl) Head(ctx context.Context, url string, opts ...func(req *fasthttp.Request)) (*fasthttp.Response, error) {
	return doRequest(ctx, c, http.MethodHead, url, nil, opts...)
}

func (c *clientImpl) Do(ctx context.Context, req *fasthttp.Request) (*fasthttp.Response, error) {
	var request = &Request{service: runtime.Project, req: req}
	request.req = req
	request.ct = filterFlags(utils.BtoS(req.Header.ContentType()))
	request.data = req.Body()
	var resp = &Response{resp: fasthttp.AcquireResponse()}
	return resp.resp, c.do(ctx, request, resp)
}

func (c *clientImpl) Get(ctx context.Context, url string, opts ...func(req *fasthttp.Request)) (*fasthttp.Response, error) {
	return doRequest(ctx, c, http.MethodGet, url, nil, opts...)
}

func (c *clientImpl) Delete(ctx context.Context, url string, opts ...func(req *fasthttp.Request)) (*fasthttp.Response, error) {
	return doRequest(ctx, c, http.MethodDelete, url, nil, opts...)
}

func (c *clientImpl) Post(ctx context.Context, url string, data interface{}, opts ...func(req *fasthttp.Request)) (*fasthttp.Response, error) {
	return doRequest(ctx, c, http.MethodPost, url, data, opts...)
}

func (c *clientImpl) PostForm(ctx context.Context, url string, val url.Values, opts ...func(req *fasthttp.Request)) (*fasthttp.Response, error) {
	var resp, err = doRequest(ctx, c, http.MethodPost, url, nil, func(req *fasthttp.Request) {
		req.Header.Set(httpx.HeaderContentType, "application/x-www-form-urlencoded")
		req.SetBodyRaw(strutil.ToBytes(val.Encode()))

		if len(opts) > 0 {
			opts[0](req)
		}
	})
	return resp, err
}

func (c *clientImpl) Put(ctx context.Context, url string, data interface{}, opts ...func(req *fasthttp.Request)) (*fasthttp.Response, error) {
	return doRequest(ctx, c, http.MethodPut, url, data, opts...)
}

func (c *clientImpl) Patch(ctx context.Context, url string, data interface{}, opts ...func(req *fasthttp.Request)) (*fasthttp.Response, error) {
	return doRequest(ctx, c, http.MethodPatch, url, data, opts...)
}

// doRequest data:[bytes|string|map|struct]
func doRequest(ctx context.Context, c *clientImpl, mth string, url string, data interface{}, opts ...func(req *fasthttp.Request)) (*fasthttp.Response, error) {
	body, err := getBodyReader(data)
	if err != nil {
		return nil, err
	}

	if ctx == nil {
		ctx = context.Background()
	}

	var req = fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	req.Header.Set(httpx.HeaderContentType, defaultContentType)
	req.Header.SetMethod(mth)
	req.Header.SetRequestURI(url)
	req.SetBodyRaw(body)
	if len(opts) > 0 {
		opts[0](req)
	}

	// Enable trace
	if c.cfg.Trace {
		ctx = (&clientTrace{}).createContext(ctx)
	}

	resp, err := c.Do(ctx, req)
	if err != nil {
		return nil, xerror.Wrap(err, mth, url)
	}

	return resp, nil
}

func filterFlags(content string) string {
	for i, char := range content {
		if char == ' ' || char == ';' {
			return content[:i]
		}
	}
	return content
}
