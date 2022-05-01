package restc

import (
	"context"
	"crypto/tls"
	"github.com/pubgo/lava/middleware"
	"net/http"
	"time"

	"github.com/pubgo/xerror"
	"github.com/valyala/fasthttp"

	"github.com/pubgo/lava/pkg/merge"
	"github.com/pubgo/lava/pkg/retry"
	"github.com/pubgo/lava/plugin"
)

type Cfg struct {
	Trace       bool
	Token       string
	Timeout     time.Duration
	RetryCount  uint32
	Proxy       bool
	Socks5      string
	Insecure    bool
	Header      map[string]string
	Middlewares []middleware.Middleware
	BasePath    string

	backoff   retry.Backoff
	tlsConfig *tls.Config
}

func (t *Cfg) Build(opts ...func(cfg *Cfg)) (_ Client, err error) {
	defer xerror.RespErr(&err)

	for i := range opts {
		opts[i](t)
	}

	if t.Timeout != 0 {
		t.backoff = retry.NewConstant(t.Timeout)
	}

	c := &http.Client{Transport: DefaultPooledTransport()}
	xerror.Panic(merge.Struct(c, t))

	//var certs []tls.Certificate
	//t.tlsConfig = &tls.Config{InsecureSkipVerify: t.Insecure, Certificates: certs}

	var middlewares []middleware.Middleware

	// 加载插件
	// 加载全局
	for _, plg := range plugin.All() {
		if plg.Middleware() == nil {
			continue
		}
		middlewares = append(middlewares, plg.Middleware())
	}

	// 加载业务自定义
	middlewares = append(middlewares, t.Middlewares...)

	var client = &clientImpl{client: &fasthttp.Client{}}
	client.do = func(ctx context.Context, req middleware.Request, resp middleware.Response) error {
		return client.client.Do(req.(*Request).req, resp.(*Response).resp)
	}
	for i := len(middlewares); i > 0; i-- {
		client.do = middlewares[i-1](client.do)
	}
	return client, nil
}

func DefaultCfg() *Cfg {
	return &Cfg{
		Timeout:    defaultHTTPTimeout,
		RetryCount: defaultRetryCount,
		backoff:    retry.NewNoop(),
	}
}
