package resty

import (
	"context"
	"crypto/tls"
	"time"

	"github.com/pubgo/funk/recovery"
	"github.com/pubgo/funk/result"
	"github.com/pubgo/funk/retry"
	"github.com/pubgo/funk/version"
	"github.com/pubgo/lava/lava"
	"github.com/valyala/fasthttp"
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

	backoff   retry.Backoff
	tlsConfig *tls.Config
}

func (t *Config) Build(mm []lava.Middleware) (r result.Result[Client]) {
	defer recovery.Result(&r)

	if t.Timeout != 0 {
		t.backoff = retry.NewConstant(t.Timeout)
	}

	var client = &clientImpl{client: &fasthttp.Client{Name: version.Project(), ReadTimeout: t.Timeout, WriteTimeout: t.Timeout}}
	client.do = func(ctx context.Context, req lava.Request) (lava.Response, error) {
		var rsp = &responseImpl{resp: fasthttp.AcquireResponse()}
		if err := client.client.Do(req.(*requestImpl).req, rsp.resp); err != nil {
			return nil, err
		}
		return rsp, nil
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
		backoff:    retry.NewNoop(),
	}
}
