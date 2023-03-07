package resty

import (
	"context"
	"crypto/tls"
	"time"

	"github.com/pubgo/funk/retry"
	"github.com/pubgo/funk/version"
	"github.com/valyala/fasthttp"

	"github.com/pubgo/lava"
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

func (t *Config) Build(mm []lava.Middleware) lava.HandlerFunc {
	if t.Timeout != 0 {
		t.backoff = retry.NewConstant(t.Timeout)
	}

	var client = &fasthttp.Client{Name: version.Project(), ReadTimeout: t.Timeout, WriteTimeout: t.Timeout}
	do := func(ctx context.Context, req lava.Request) (lava.Response, error) {
		var rsp = &responseImpl{resp: fasthttp.AcquireResponse()}
		if err := client.Do(req.(*requestImpl).req, rsp.resp); err != nil {
			return nil, err
		}
		return rsp, nil
	}

	do = lava.Chain(mm...)(do)
	return do
}

func DefaultCfg() *Config {
	return &Config{
		Timeout:    defaultHTTPTimeout,
		RetryCount: defaultRetryCount,
		backoff:    retry.NewNoop(),
	}
}
