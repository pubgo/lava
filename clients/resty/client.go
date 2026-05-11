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

	"github.com/pubgo/lava/v2/core/metrics"
	"github.com/pubgo/lava/v2/internal/middlewares/middleware_accesslog"
	"github.com/pubgo/lava/v2/internal/middlewares/middleware_metric"
	"github.com/pubgo/lava/v2/internal/middlewares/middleware_recovery"
	"github.com/pubgo/lava/v2/internal/middlewares/middleware_serviceinfo"
	"github.com/pubgo/lava/v2/lava"
)

// Params 客户端参数结构
// Log: 日志记录器
// Metric: 指标收集器
type Params struct {
	Log    log.Logger
	Metric metrics.Metric
}

// New 创建一个新的 HTTP 客户端
// cfg: 客户端配置
// p: 客户端参数
// mm: 自定义中间件
// 返回: 初始化后的客户端实例
func New(cfg *Config, p Params, mm ...lava.Middleware) *Client {
	cfg = config.MergeR(DefaultCfg(), cfg).Unwrap()
	middlewares := make(lava.Middlewares, 0, 4+len(mm))
	middlewares = append(middlewares,
		middleware_serviceinfo.New(),
		middleware_metric.New(p.Metric),
		middleware_accesslog.New(p.Log.WithFields(log.Fields{"service": cfg.ServiceName})),
		middleware_recovery.New(),
	)
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

// Client 是 IClient 的实现，提供 HTTP 客户端功能
type Client struct {
	do            lava.HandlerFunc // 处理函数
	log           log.Logger       // 日志记录器
	cfg           *Config          // 客户端配置
	baseUrl       *url.URL         // 基础 URL
	backoff       retry.Backoff    // 重试策略
	pathTemplates sync.Map         // 路径模板缓存
}

// Do 发送 HTTP 请求
// ctx: 上下文
// req: 请求对象
// 返回: 响应结果
func (c *Client) Do(ctx context.Context, req *Request) (r result.Result[*Response]) {
	defer result.Recovery(&r)

	if doRequest(c, req).ValueTo(&req.req).Throw(&r) {
		return r
	}

	request := &requestImpl{service: c.cfg.ServiceName, req: req}
	resp, err := c.do(ctx, request)
	if err != nil {
		return r.WithErr(err)
	}

	return r.WithValue(&Response{resp: resp.(*responseImpl).resp})
}
