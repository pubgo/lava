package resty

import (
	"fmt"
	"net"
	"time"

	"github.com/pubgo/funk/version"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttpproxy"
	"golang.org/x/net/http/httpproxy"
)

type Config struct {
	BaseUrl              string            `yaml:"base_url"`
	ServiceName          string            `yaml:"service_name"`
	DefaultHeader        map[string]string `yaml:"default_header"`
	DefaultContentType   string            `yaml:"default_content_type"`
	DefaultRetryCount    uint32            `yaml:"default_retry_count"`
	DefaultRetryInterval time.Duration     `yaml:"default_retry_interval"`
	BasicToken           string            `yaml:"basic_token"`
	JwtToken             string            `yaml:"jwt_token"`

	EnableProxy               bool          `yaml:"enable_proxy"`
	EnableAuth                bool          `yaml:"enable_auth"`
	DialTimeout               time.Duration `yaml:"dial_timeout"`
	ReadTimeout               time.Duration `yaml:"read_timeout"`
	WriteTimeout              time.Duration `yaml:"write_timeout"`
	MaxConnsPerHost           int           `yaml:"max_conns_per_host"`
	MaxIdleConnDuration       time.Duration `yaml:"max_idle_conn_duration"`
	MaxIdemponentCallAttempts int           `yaml:"max_idemponent_call_attempts"`
	ReadBufferSize            int           `yaml:"read_buffer_size"`
	WriteBufferSize           int           `yaml:"write_buffer_size"`
	MaxResponseBodySize       int           `yaml:"max_response_body_size"`
}

func (t *Config) Build() *fasthttp.Client {
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

	if t.EnableProxy && httpproxy.FromEnvironment() != nil {
		client.Dial = fasthttpproxy.FasthttpProxyHTTPDialerTimeout(t.DialTimeout)
	} else {
		client.Dial = func(addr string) (net.Conn, error) {
			return fasthttp.DialTimeout(addr, t.DialTimeout)
		}
	}

	return client
}

func DefaultCfg() *Config {
	return &Config{
		DialTimeout:               defaultHTTPTimeout,
		ReadTimeout:               defaultTimeout,
		WriteTimeout:              defaultTimeout,
		DefaultRetryCount:         defaultRetryCount,
		DefaultRetryInterval:      defaultRetryInterval,
		MaxConnsPerHost:           512,
		MaxIdleConnDuration:       10 * time.Second,
		MaxIdemponentCallAttempts: 5,
		ReadBufferSize:            4096,
		WriteBufferSize:           4096,
		MaxResponseBodySize:       2 * 1024 * 1024,
	}
}
