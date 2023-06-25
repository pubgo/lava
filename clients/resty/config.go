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
	BaseUrl string `yaml:"base_url"`

	Timeout                   time.Duration     `yaml:"timeout"`
	ReadTimeout               time.Duration     `yaml:"read_timeout"`
	WriteTimeout              time.Duration     `yaml:"write_timeout"`
	RetryCount                uint32            `yaml:"retry_count"`
	Proxy                     bool              `yaml:"proxy"`
	Socks5                    string            `yaml:"socks5"`
	Insecure                  bool              `yaml:"insecure"`
	Header                    map[string]string `yaml:"header"`
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

func (t *Config) Build() *fasthttp.Client {
	if t.Timeout != 0 {
		t.backoff = retry.NewConstant(t.Timeout)
	}

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
