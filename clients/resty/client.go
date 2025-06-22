package resty

import (
	"context"
	"net/url"
	"sync"

	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/config"
	"github.com/pubgo/funk/log"
	"github.com/pubgo/funk/retry"
	"github.com/pubgo/funk/v2/result"
	"github.com/valyala/fasthttp"

	"github.com/pubgo/lava/core/metrics"
	"github.com/pubgo/lava/internal/middlewares/middleware_accesslog"
	"github.com/pubgo/lava/internal/middlewares/middleware_metric"
	"github.com/pubgo/lava/internal/middlewares/middleware_recovery"
	"github.com/pubgo/lava/internal/middlewares/middleware_service_info"
	"github.com/pubgo/lava/lava"
)

type Params struct {
	Log    log.Logger
	Metric metrics.Metric
}

func New(cfg *Config, p Params, mm ...lava.Middleware) *Client {
	cfg = config.MergeR(DefaultCfg(), cfg).Unwrap()
	middlewares := lava.Middlewares{
		middleware_service_info.New(),
		middleware_metric.New(p.Metric),
		middleware_accesslog.New(p.Log.WithFields(log.Map{"service": cfg.ServiceName})),
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

	return &Client{
		do:      lava.Chain(middlewares...).Middleware(do(cfg)),
		log:     p.Log,
		cfg:     cfg,
		baseUrl: assert.Must1(url.Parse(cfg.BaseUrl)),
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
	defer result.RecoveryErr(&r)

	reqErr := doRequest(c, req)
	if reqErr.IsErr() {
		return r.WithErr(reqErr.Err())
	}

	req.req = reqErr.Unwrap()

	request := &requestImpl{service: c.cfg.ServiceName, req: req}
	resp, err := c.do(ctx, request)
	if err != nil {
		return r.WithErr(err)
	}

	return r.WithValue(resp.(*responseImpl).resp)
}
