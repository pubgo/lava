package resty

import (
	"context"
	"net/url"
	"sync"

	"github.com/pubgo/funk/v2/assert"
	"github.com/pubgo/funk/v2/config"
	"github.com/pubgo/funk/v2/log"
	"github.com/pubgo/funk/v2/result"
	"github.com/pubgo/funk/v2/retry"
	"github.com/valyala/fasthttp"

	"github.com/pubgo/lava/v2/core/metrics"
	"github.com/pubgo/lava/v2/internal/middlewares/middleware_accesslog"
	"github.com/pubgo/lava/v2/internal/middlewares/middleware_metric"
	"github.com/pubgo/lava/v2/internal/middlewares/middleware_recovery"
	"github.com/pubgo/lava/v2/internal/middlewares/middleware_serviceinfo"
	"github.com/pubgo/lava/v2/lava"
)

type Params struct {
	Log    log.Logger
	Metric metrics.Metric
}

func New(cfg *Config, p Params, mm ...lava.Middleware) *Client {
	cfg = config.MergeR(DefaultCfg(), cfg).Unwrap()
	middlewares := lava.Middlewares{
		middleware_serviceinfo.New(),
		middleware_metric.New(p.Metric),
		middleware_accesslog.New(p.Log.WithFields(log.Fields{"service": cfg.ServiceName})),
		middleware_recovery.New(),
	}
	middlewares = append(middlewares, mm...)

	var backoff retry.Backoff
	if cfg.DefaultRetryInterval > 0 {
		backoff = retry.NewConstant(cfg.DefaultRetryInterval)
	}

	if cfg.DefaultRetryCount > 0 {
		backoff = retry.WithMaxRetries(cfg.DefaultRetryCount, backoff)
	}

	handler := do(cfg)
	handler = lava.Chain(middlewares...).Middleware(handler)

	baseUrl := assert.Must1(url.Parse(cfg.BaseUrl))

	return &Client{
		do:      handler,
		log:     p.Log,
		cfg:     cfg,
		baseUrl: baseUrl,
		backoff: backoff,
	}
}

var _ IClient = (*Client)(nil)

// Client is the IClient implementation
type Client struct {
	do            lava.HandlerFunc
	log           log.Logger
	cfg           *Config
	baseUrl       *url.URL
	backoff       retry.Backoff
	pathTemplates sync.Map
}

func (c *Client) Do(ctx context.Context, req *Request) (r result.Result[*fasthttp.Response]) {
	defer result.Recovery(&r)

	if doRequest(c, req).ValueTo(&req.req).Throw(&r) {
		return r
	}

	request := &requestImpl{service: c.cfg.ServiceName, req: req}
	resp, err := c.do(ctx, request)
	if err != nil {
		return r.WithErr(err)
	}

	return r.WithValue(resp.(*responseImpl).resp)
}
