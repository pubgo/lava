package restc

import (
	"crypto/tls"
	"runtime"
	"strings"
	"time"

	"github.com/pubgo/lug/pkg/retry"
	"github.com/pubgo/x/merge"
	"github.com/pubgo/xerror"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttpproxy"
)

type Cfg struct {
	Token                         string
	keepAlive                     bool
	Timeout                       time.Duration
	RetryCount                    uint64
	Name                          string
	NoDefaultUserAgentHeader      bool
	DialDualStack                 bool
	MaxConnsPerHost               int
	MaxIdleConnDuration           time.Duration
	MaxConnDuration               time.Duration
	MaxIdemponentCallAttempts     int
	ReadBufferSize                int
	WriteBufferSize               int
	ReadTimeout                   time.Duration
	WriteTimeout                  time.Duration
	MaxResponseBodySize           int
	DisableHeaderNamesNormalizing bool
	DisablePathNormalizing        bool
	MaxConnWaitTimeout            time.Duration
	Socks5Proxy                   string
	CertPath                      string
	KeyPath                       string
	Insecure                      bool
	Header                        map[string][]string
	retryIf                       fasthttp.RetryIfFunc
	backoff                       retry.Backoff
	tlsConfig                     *tls.Config
	dial                          fasthttp.DialFunc
	middles                       []Middleware
}

func (t Cfg) Build(opts ...func(cfg *Cfg)) (_ Client, err error) {
	defer xerror.RespErr(&err)

	for i := range opts {
		opts[i](&t)
	}

	c := &fasthttp.Client{}
	xerror.Panic(merge.CopyStruct(c, t))

	if t.Socks5Proxy != "" {
		if !strings.Contains(t.Socks5Proxy, "://") {
			t.Socks5Proxy = "socks5://" + t.Socks5Proxy
		}
		c.Dial = fasthttpproxy.FasthttpSocksDialer(t.Socks5Proxy)
	} else {
		c.Dial = fasthttpproxy.FasthttpProxyHTTPDialerTimeout(defaultHTTPTimeout)
	}
	if t.dial != nil {
		c.Dial = t.dial
	}

	var certs []tls.Certificate
	if t.CertPath != "" && t.KeyPath != "" {
		c, err := tls.LoadX509KeyPair(t.CertPath, t.KeyPath)
		xerror.Panic(err)
		certs = append(certs, c)
	}
	c.TLSConfig = &tls.Config{InsecureSkipVerify: t.Insecure, Certificates: certs}

	var requestHeader fasthttp.RequestHeader
	requestHeader.SetMethod(fasthttp.MethodGet)
	requestHeader.SetContentType(defaultContentType)
	requestHeader.Set(fasthttp.HeaderConnection, "close")
	if t.keepAlive {
		requestHeader.Set(fasthttp.HeaderConnection, "keep-alive")
	}

	if t.Token != "" {
		requestHeader.Set(fasthttp.HeaderAuthorization, t.Token)
	}

	if t.Header != nil {
		for k, v := range t.Header {
			for i := range v {
				requestHeader.Add(k, v[i])
			}
		}
	}

	var client = &client{client: c, defaultHeader: &requestHeader}
	client.do = doFunc(client)
	for i := len(t.middles); i > 0; i-- {
		client.do = t.middles[i-1](client.do)
	}
	return client, nil
}

func DefaultCfg() Cfg {
	return Cfg{
		DialDualStack:       true,
		Timeout:             defaultHTTPTimeout,
		ReadTimeout:         defaultHTTPTimeout,
		WriteTimeout:        defaultHTTPTimeout,
		RetryCount:          defaultRetryCount,
		backoff:             retry.NewNoop(),
		MaxIdleConnDuration: 90 * time.Second,
		MaxConnWaitTimeout:  30 * time.Second,
		MaxConnsPerHost:     runtime.GOMAXPROCS(0) + 1, // http client缓存数量
	}
}
