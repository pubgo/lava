package resty

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"time"

	"github.com/pubgo/funk/retry"
	"github.com/pubgo/funk/version"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttpproxy"

	"github.com/pubgo/lava"
)

func (c *clientConfig) Initial() {
	c.DialDualStack = false
	c.MaxConnsPerHost = 512
	c.MaxIdleConnDuration = 10 * time.Second
	c.MaxIdemponentCallAttempts = 5
	c.ReadBufferSize = 4096
	c.WriteBufferSize = 4096
	c.ReadTimeout = 10 * time.Second
	c.WriteTimeout = 10 * time.Second
	c.MaxResponseBodySize = 2 * 1024 * 1024
}

type clientConfig struct {
	DialDualStack             bool          `json:"dialDualStack"`
	MaxConnsPerHost           int           `josn:"maxConnsPerHost"`
	MaxIdleConnDuration       time.Duration `json:"maxIdleConnDuration"`
	MaxIdemponentCallAttempts int           `json:"maxIdemponentCallAttempts"`
	ReadBufferSize            int           `json:"readBufferSize"`
	WriteBufferSize           int           `json:"writeBufferSize"`
	ReadTimeout               time.Duration `json:"readTimeout"`
	WriteTimeout              time.Duration `json:"writeTimeout"`
	MaxResponseBodySize       int           `json:"maxResponseBodySize"`
}

type Config struct {
	Trace                     bool              `yaml:"trace"`
	Timeout                   time.Duration     `yaml:"timeout"`
	ReadTimeout               time.Duration     `yaml:"read_timeout"`
	WriteTimeout              time.Duration     `yaml:"write_timeout"`
	RetryCount                uint32            `yaml:"retry_count"`
	Proxy                     bool              `yaml:"proxy"`
	Socks5                    string            `yaml:"socks5"`
	Insecure                  bool              `yaml:"insecure"`
	Header                    map[string]string `yaml:"header"`
	BaseUrl                   string            `yaml:"base_url"`
	TargetService             string            `yaml:"target_service"`
	Token                     string            `yaml:"token"`
	JwtToken                  string            `yaml:"jwt_token"`
	UserAgent                 string            `yaml:"user_agent"`
	ContentType               string            `yaml:"content_type"`
	Accept                    string            `yaml:"accept"`
	Debug                     bool              `yaml:"debug"`
	Authentication            bool              `yaml:"authentication"`
	GzipRequest               bool              `yaml:"gzip_request"`
	MaxConnsPerHost           int               `yaml:"max_conns_per_host"`
	MaxIdleConnDuration       time.Duration     `yaml:"max_idle_conn_duration"`
	MaxIdemponentCallAttempts int               `yaml:"max_idemponent_call_attempts"`
	ReadBufferSize            int               `yaml:"read_buffer_size"`
	WriteBufferSize           int               `yaml:"write_buffer_size"`
	MaxResponseBodySize       int               `yaml:"max_response_body_size"`

	proxy     string // set to all requests
	backoff   retry.Backoff
	tlsConfig *tls.Config
}

func (t *Config) Build(mm []lava.Middleware) lava.HandlerFunc {
	if t.Timeout != 0 {
		t.backoff = retry.NewConstant(t.Timeout)
	}

	var client = &fasthttp.Client{
		Name:                      fmt.Sprintf("%s: %s", version.Project(), version.Version()),
		ReadTimeout:               t.Timeout,
		WriteTimeout:              t.Timeout,
		NoDefaultUserAgentHeader:  true,
		MaxIdemponentCallAttempts: 5,
	}

	if t.proxy != "" {
		client.Dial = fasthttpproxy.FasthttpHTTPDialer(t.proxy)
	} else {
		client.Dial = func(addr string) (net.Conn, error) {
			return fasthttp.DialTimeout(addr, t.Timeout)
		}
	}

	do := func(ctx context.Context, req lava.Request) (lava.Response, error) {
		req.Header().SetUserAgent(fmt.Sprintf("%s: %s", version.Project(), version.Version()))

		var err error
		var resp = fasthttp.AcquireResponse()
		deadline, ok := ctx.Deadline()
		if ok {
			err = client.DoDeadline(req.(*requestImpl).req, resp, deadline)
		} else {
			err = client.Do(req.(*requestImpl).req, resp)
		}

		if err != nil {
			return nil, err
		}

		return &responseImpl{resp: resp}, nil
	}

	do = lava.Chain(mm...)(do)
	return do
}

func DefaultCfg() *Config {
	return &Config{
		Timeout:      defaultHTTPTimeout,
		ReadTimeout:  6 * time.Second,
		WriteTimeout: 6 * time.Second,
		RetryCount:   defaultRetryCount,
		backoff:      retry.NewNoop(),
	}
}
