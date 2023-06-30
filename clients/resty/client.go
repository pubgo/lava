package resty

import (
	"context"
	"net/http"
	"net/url"
	"sync"

	"github.com/pubgo/funk/config"
	"github.com/pubgo/funk/convert"
	"github.com/pubgo/funk/log"
	"github.com/pubgo/funk/recovery"
	"github.com/pubgo/funk/result"
	"github.com/valyala/fasthttp"

	"github.com/pubgo/lava/internal/middlewares/middleware_accesslog"
	"github.com/pubgo/lava/internal/middlewares/middleware_metric"
	"github.com/pubgo/lava/internal/middlewares/middleware_recovery"
	"github.com/pubgo/lava/internal/middlewares/middleware_service_info"
	"github.com/pubgo/lava/lava"
)

type Params struct {
	Log                 log.Logger
	MetricMiddleware    *middleware_metric.MetricMiddleware
	AccessLogMiddleware *middleware_accesslog.LogMiddleware
}

func New(cfg *Config, p Params, mm ...lava.Middleware) *Client {
	cfg = config.MergeR(DefaultCfg(), cfg).Unwrap()
	middlewares := lava.Middlewares{
		middleware_service_info.New(), p.MetricMiddleware, p.AccessLogMiddleware, middleware_recovery.New(),
	}
	middlewares = append(middlewares, mm...)

	return &Client{
		do:  lava.Chain(middlewares...).Middleware(do(cfg)),
		log: p.Log,
	}
}

var _ IClient = (*Client)(nil)

// Client is the Client implementation
type Client struct {
	do            lava.HandlerFunc
	log           log.Logger
	cfg           *Config
	baseUrl       *url.URL
	pathTemplates sync.Map
}

func (c *Client) Do(ctx context.Context, req *Request) (r result.Result[*fasthttp.Response]) {
	defer recovery.Result(&r)
	defer fasthttp.ReleaseRequest(req.req)

	request := &requestImpl{service: c.cfg.ServiceName, req: req.req}
	request.ct = filterFlags(convert.BtoS(req.Header.ContentType()))
	request.data = req.Body()
	resp, err := c.do(ctx, request)
	if err != nil {
		return r.WithErr(err)
	}

	return r.WithVal(resp.(*responseImpl).resp)
}

func (c *Client) Head(ctx context.Context, req *Request) result.Result[*fasthttp.Response] {
	return doRequest(ctx, c, http.MethodHead, req)
}

func (c *Client) Get(ctx context.Context, req *Request) result.Result[*fasthttp.Response] {
	return doRequest(ctx, c, http.MethodGet, req)
}

func (c *Client) Delete(ctx context.Context, req *Request) result.Result[*fasthttp.Response] {
	return doRequest(ctx, c, http.MethodDelete, req)
}

func (c *Client) Post(ctx context.Context, data interface{}, req *Request) result.Result[*fasthttp.Response] {
	return doRequest(ctx, c, http.MethodPost, req)
}

func (c *Client) PostForm(ctx context.Context, val url.Values, req *Request) result.Result[*fasthttp.Response] {
	req.SetClient(c)
	req.SetContentType("application/x-www-form-urlencoded")
	return req.Post(ctx, val)
}

func (c *Client) Put(ctx context.Context, data interface{}, req *Request) result.Result[*fasthttp.Response] {
	return doRequest(ctx, c, http.MethodPut, req)
}

func (c *Client) Patch(ctx context.Context, data interface{}, req *Request) result.Result[*fasthttp.Response] {
	return doRequest(ctx, c, http.MethodPatch, req)
}
