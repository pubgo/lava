package resty

import (
	"context"
	"crypto/tls"
	"net/http"
	"time"

	"github.com/pubgo/funk/recovery"
	"github.com/pubgo/funk/result"
	"github.com/valyala/fasthttp"

	"github.com/pubgo/lava/pkg/merge"
	retry2 "github.com/pubgo/lava/pkg/retry"
	"github.com/pubgo/lava/service"
)

const (
	defaultRetryCount  = 1
	defaultHTTPTimeout = 2 * time.Second
	defaultContentType = "application/json"
)

type Config struct {
	Trace      bool              `yaml:"trace"`
	Token      string            `yaml:"token"`
	Timeout    time.Duration     `yaml:"timeout"`
	RetryCount uint32            `yaml:"retry-count"`
	Proxy      bool              `yaml:"proxy"`
	Socks5     string            `yaml:"socks5"`
	Insecure   bool              `yaml:"insecure"`
	Header     map[string]string `yaml:"header"`
	BasePath   string            `yaml:"base-path"`

	backoff   retry2.Backoff
	tlsConfig *tls.Config
}

func (t *Config) Build(mm []service.Middleware) (r result.Result[Client]) {
	defer recovery.Result(&r)

	if t.Timeout != 0 {
		t.backoff = retry2.NewConstant(t.Timeout)
	}

	c := &http.Client{Transport: DefaultPooledTransport()}
	merge.Struct(c, t).Unwrap()

	var client = &clientImpl{client: &fasthttp.Client{}}
	client.do = func(ctx context.Context, req service.Request, resp service.Response) error {
		return client.client.Do(req.(*Request).req, resp.(*Response).resp)
	}

	for i := len(mm); i > 0; i-- {
		client.do = mm[i-1](client.do)
	}
	return r.WithVal(client)
}

func DefaultCfg() *Config {
	return &Config{
		Timeout:    defaultHTTPTimeout,
		RetryCount: defaultRetryCount,
		backoff:    retry2.NewNoop(),
	}
}
