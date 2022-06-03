package restc

import (
	"context"
	"crypto/tls"
	"net/http"
	"time"

	"github.com/pubgo/xerror"
	"github.com/valyala/fasthttp"

	"github.com/pubgo/lava/logging/log_middleware"
	"github.com/pubgo/lava/middleware"
	"github.com/pubgo/lava/pkg/merge"
	"github.com/pubgo/lava/pkg/retry"
)

type Cfg struct {
	Trace       bool              `yaml:"trace"`
	Token       string            `yaml:"token"`
	Timeout     time.Duration     `yaml:"timeout"`
	RetryCount  uint32            `yaml:"retry-count"`
	Proxy       bool              `yaml:"proxy"`
	Socks5      string            `yaml:"socks5"`
	Insecure    bool              `yaml:"insecure"`
	Header      map[string]string `yaml:"header"`
	Middlewares []string          `yaml:"middlewares"`
	BasePath    string            `yaml:"base-path"`
	backoff     retry.Backoff
	tlsConfig   *tls.Config
}

func (t *Cfg) Build(opts ...func(cfg *Cfg)) (_ Client, err error) {
	defer xerror.RecoverErr(&err)

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
	for _, plg := range t.Middlewares {
		middlewares = append(middlewares, middleware.Get(plg))
	}

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
		Timeout:     defaultHTTPTimeout,
		RetryCount:  defaultRetryCount,
		backoff:     retry.NewNoop(),
		Middlewares: []string{log_middleware.Name},
	}
}
