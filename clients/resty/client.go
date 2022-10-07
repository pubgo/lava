package resty

import (
	"context"
	"net/http"
	"net/url"

	"github.com/pubgo/funk/result"
	"github.com/pubgo/x/strutil"
	"github.com/valyala/fasthttp"

	"github.com/pubgo/lava/core/projectinfo"
	"github.com/pubgo/lava/core/requestid"
	"github.com/pubgo/lava/core/runmode"
	"github.com/pubgo/lava/logging"
	"github.com/pubgo/lava/logging/logmiddleware"
	"github.com/pubgo/lava/pkg/httpx"
	"github.com/pubgo/lava/pkg/merge"
	"github.com/pubgo/lava/pkg/utils"
	"github.com/pubgo/lava/service"
)

func New(cfg *Config, log *logging.Logger, mm ...service.Middleware) Client {
	cfg = merge.Copy(DefaultCfg(), cfg).Unwrap()
	return cfg.Build(append([]service.Middleware{
		logmiddleware.Middleware(log),
		requestid.Middleware(),
		projectinfo.Middleware(),
	}, mm...)).Unwrap()
}

var _ Client = (*clientImpl)(nil)

// clientImpl is the Client implementation
type clientImpl struct {
	client *fasthttp.Client
	cfg    Config
	do     service.HandlerFunc
}

func (c *clientImpl) Head(ctx context.Context, url string, opts ...func(req *fasthttp.Request)) result.Result[*fasthttp.Response] {
	return doRequest(ctx, c, http.MethodHead, url, nil, opts...)
}

func (c *clientImpl) Do(ctx context.Context, req *fasthttp.Request) (r result.Result[*fasthttp.Response]) {
	var request = &Request{service: runmode.Project, req: req}
	request.req = req
	request.ct = filterFlags(utils.BtoS(req.Header.ContentType()))
	request.data = req.Body()
	var resp = &Response{resp: fasthttp.AcquireResponse()}
	if err := c.do(ctx, request, resp); err != nil {
		return r.WithErr(err)
	}
	return r.WithVal(resp.resp)
}

func (c *clientImpl) Get(ctx context.Context, url string, opts ...func(req *fasthttp.Request)) result.Result[*fasthttp.Response] {
	return doRequest(ctx, c, http.MethodGet, url, nil, opts...)
}

func (c *clientImpl) Delete(ctx context.Context, url string, opts ...func(req *fasthttp.Request)) result.Result[*fasthttp.Response] {
	return doRequest(ctx, c, http.MethodDelete, url, nil, opts...)
}

func (c *clientImpl) Post(ctx context.Context, url string, data interface{}, opts ...func(req *fasthttp.Request)) result.Result[*fasthttp.Response] {
	return doRequest(ctx, c, http.MethodPost, url, data, opts...)
}

func (c *clientImpl) PostForm(ctx context.Context, url string, val url.Values, opts ...func(req *fasthttp.Request)) result.Result[*fasthttp.Response] {
	return doRequest(ctx, c, http.MethodPost, url, nil, func(req *fasthttp.Request) {
		req.Header.Set(httpx.HeaderContentType, "application/x-www-form-urlencoded")
		req.SetBodyRaw(strutil.ToBytes(val.Encode()))

		if len(opts) > 0 {
			opts[0](req)
		}
	})
}

func (c *clientImpl) Put(ctx context.Context, url string, data interface{}, opts ...func(req *fasthttp.Request)) result.Result[*fasthttp.Response] {
	return doRequest(ctx, c, http.MethodPut, url, data, opts...)
}

func (c *clientImpl) Patch(ctx context.Context, url string, data interface{}, opts ...func(req *fasthttp.Request)) result.Result[*fasthttp.Response] {
	return doRequest(ctx, c, http.MethodPatch, url, data, opts...)
}

// doRequest data:[bytes|string|map|struct]
func doRequest(ctx context.Context, c *clientImpl, mth string, url string, data interface{}, opts ...func(req *fasthttp.Request)) (r result.Result[*fasthttp.Response]) {
	body, err := getBodyReader(data)
	if err != nil {
		return r.WithErr(err)
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

	return c.Do(ctx, req)
}

func filterFlags(content string) string {
	for i, char := range content {
		if char == ' ' || char == ';' {
			return content[:i]
		}
	}
	return content
}
