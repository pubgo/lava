package restc

import (
	"context"
	"crypto/tls"
	"net/http"
	"time"

	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/recovery"
	"github.com/pubgo/funk/result"
	"github.com/valyala/fasthttp"

	"github.com/pubgo/lava/logging"
	"github.com/pubgo/lava/pkg/merge"
	retry2 "github.com/pubgo/lava/pkg/retry"
	"github.com/pubgo/lava/service"
)

type Config struct {
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
	backoff     retry2.Backoff
	tlsConfig   *tls.Config
}

func (t *Config) Build(mm map[string]service.Middleware) (r result.Result[Client]) {
	defer recovery.Result(&r)

	if t.Timeout != 0 {
		t.backoff = retry2.NewConstant(t.Timeout)
	}

	c := &http.Client{Transport: DefaultPooledTransport()}
	merge.Struct(c, t).Unwrap()

	var middlewares []service.Middleware

	// 加载插件
	for _, m := range t.Middlewares {
		assert.If(mm[m] == nil, "middleware %s not found", m)
		middlewares = append(middlewares, mm[m])
	}

	var client = &clientImpl{client: &fasthttp.Client{}}
	client.do = func(ctx context.Context, req service.Request, resp service.Response) error {
		return client.client.Do(req.(*Request).req, resp.(*Response).resp)
	}
	for i := len(middlewares); i > 0; i-- {
		client.do = middlewares[i-1](client.do)
	}
	return r.WithVal(client)
}

func DefaultCfg() *Config {
	return &Config{
		Timeout:     defaultHTTPTimeout,
		RetryCount:  defaultRetryCount,
		backoff:     retry2.NewNoop(),
		Middlewares: []string{logging.Name},
	}
}
