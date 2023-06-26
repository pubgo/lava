package resty

import (
	"crypto/tls"
	"fmt"
	"net"
	"time"

	"github.com/pubgo/funk/retry"
	"github.com/pubgo/funk/version"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttpproxy"
	"golang.org/x/net/http/httpproxy"
)

type Config struct {
	BaseUrl            string            `yaml:"base_url"`
	DefaultHeader      map[string]string `yaml:"default_header"`
	DefaultContentType string            `yaml:"default_content_type"`
	RetryCount         uint32            `yaml:"retry_count"`
	Proxy              bool              `yaml:"proxy"`
	Socks5             string            `yaml:"socks5"`
	TargetService      string            `yaml:"target_service"`
	BasicToken         string            `yaml:"token"`
	JwtToken           string            `yaml:"jwt_token"`
	Debug              bool              `yaml:"debug"`
	GzipRequest        bool              `yaml:"gzip_request"`

	Timeout                   time.Duration `yaml:"timeout"`
	ReadTimeout               time.Duration `yaml:"read_timeout"`
	WriteTimeout              time.Duration `yaml:"write_timeout"`
	MaxConnsPerHost           int           `yaml:"max_conns_per_host"`
	MaxIdleConnDuration       time.Duration `yaml:"max_idle_conn_duration"`
	MaxIdemponentCallAttempts int           `yaml:"max_idemponent_call_attempts"`
	ReadBufferSize            int           `yaml:"read_buffer_size"`
	WriteBufferSize           int           `yaml:"write_buffer_size"`
	MaxResponseBodySize       int           `yaml:"max_response_body_size"`

	backoff   retry.Backoff
	tlsConfig *tls.Config
}

func (t *Config) Build() *fasthttp.Client {
	if t.Timeout != 0 {
		t.backoff = retry.NewConstant(t.Timeout)
	}

	t.backoff.Next()

	client := &fasthttp.Client{
		Name:                      fmt.Sprintf("%s: %s", version.Project(), version.Version()),
		ReadTimeout:               t.ReadTimeout,
		WriteTimeout:              t.WriteTimeout,
		NoDefaultUserAgentHeader:  true,
		MaxConnsPerHost:           t.MaxConnsPerHost,
		MaxIdleConnDuration:       t.MaxIdleConnDuration,
		MaxIdemponentCallAttempts: t.MaxIdemponentCallAttempts,
		ReadBufferSize:            t.ReadBufferSize,
		WriteBufferSize:           t.WriteBufferSize,
		MaxResponseBodySize:       t.MaxResponseBodySize,
	}

	if t.Proxy && httpproxy.FromEnvironment() != nil {
		client.Dial = fasthttpproxy.FasthttpProxyHTTPDialerTimeout(t.Timeout)
	} else {
		client.Dial = func(addr string) (net.Conn, error) {
			return fasthttp.DialTimeout(addr, t.Timeout)
		}
	}

	return client
}

func DefaultCfg() *Config {
	return &Config{
		backoff: retry.NewNoop(),

		Timeout:                   defaultHTTPTimeout,
		ReadTimeout:               10 * time.Second,
		WriteTimeout:              10 * time.Second,
		RetryCount:                defaultRetryCount,
		MaxConnsPerHost:           512,
		MaxIdleConnDuration:       10 * time.Second,
		MaxIdemponentCallAttempts: 5,
		ReadBufferSize:            4096,
		WriteBufferSize:           4096,
		MaxResponseBodySize:       2 * 1024 * 1024,
	}
}
