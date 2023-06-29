package resty

import (
	"context"
	"net/http"
	"net/url"

	"github.com/pubgo/funk/config"
	"github.com/pubgo/funk/convert"
	"github.com/pubgo/funk/log"
	"github.com/pubgo/funk/recovery"
	"github.com/pubgo/funk/result"
	"github.com/pubgo/funk/version"
	"github.com/valyala/fasthttp"

	"github.com/pubgo/lava/internal/middlewares/middleware_accesslog"
	"github.com/pubgo/lava/internal/middlewares/middleware_metric"
	"github.com/pubgo/lava/internal/middlewares/middleware_recovery"
	"github.com/pubgo/lava/internal/middlewares/middleware_service_info"
	"github.com/pubgo/lava/lava"
	"github.com/pubgo/lava/pkg/httputil"
)

type Params struct {
	Log                 log.Logger
	MetricMiddleware    *middleware_metric.MetricMiddleware
	AccessLogMiddleware *middleware_accesslog.LogMiddleware
}

func New(cfg *Config, p Params, mm ...lava.Middleware) Client {
	cfg = config.MergeR(DefaultCfg(), cfg).Unwrap()
	middlewares := lava.Middlewares{
		middleware_service_info.New(), p.MetricMiddleware, p.AccessLogMiddleware, middleware_recovery.New(),
	}
	middlewares = append(middlewares, mm...)

	return &clientImpl{
		do:  lava.Chain(middlewares...).Middleware(do(cfg)),
		log: p.Log,
	}
}

var _ Client = (*clientImpl)(nil)

// clientImpl is the Client implementation
type clientImpl struct {
	do      lava.HandlerFunc
	log     log.Logger
	cfg     *Config
	baseUrl *url.URL
}

func (c *clientImpl) Do(ctx context.Context, req *Request) (r result.Result[*fasthttp.Response]) {
	defer recovery.Result(&r)
	defer fasthttp.ReleaseRequest(req.req)

	request := &requestImpl{service: version.Project(), req: req.req}
	request.ct = filterFlags(convert.BtoS(req.Header.ContentType()))
	request.data = req.Body()
	resp, err := c.do(ctx, request)
	if err != nil {
		return r.WithErr(err)
	}

	return r.WithVal(resp.(*responseImpl).resp)
}

func (c *clientImpl) Head(ctx context.Context, url string, opts ...func(req *fasthttp.Request)) result.Result[*fasthttp.Response] {
	return doRequest(ctx, c, http.MethodHead, url, nil, opts...)
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
		req.Header.Set(httputil.HeaderContentType, "application/x-www-form-urlencoded")
		req.SetBodyRaw(convert.StoB(val.Encode()))

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
