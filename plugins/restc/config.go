package restc

import (
	"crypto/tls"
	"runtime"
	"strings"
	"time"

	"github.com/pubgo/x/merge"
	"github.com/pubgo/xerror"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttpproxy"

	"github.com/pubgo/lug/pkg/retry"
	"github.com/pubgo/lug/types"
)

type Cfg struct {
	Token                         string
	KeepAlive                     bool
	Timeout                       time.Duration
	RetryCount                    uint32
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
	Proxy                         bool
	Socks5                        string
	CertPath                      string
	KeyPath                       string
	Insecure                      bool
	Header                        map[string]string
	retryIf                       fasthttp.RetryIfFunc
	backoff                       retry.Backoff
	tlsConfig                     *tls.Config
	dial                          fasthttp.DialFunc
	middles                       []types.Middleware
}

func (t Cfg) Build(opts ...func(cfg *Cfg)) (_ Client, err error) {
	defer xerror.RespErr(&err)

	for i := range opts {
		opts[i](&t)
	}

	c := &fasthttp.Client{}
	xerror.Panic(merge.CopyStruct(c, t))

	if t.Proxy {
		if t.Socks5 != "" {
			if !strings.Contains(t.Socks5, "://") {
				t.Socks5 = "socks5://" + t.Socks5
			}
			c.Dial = fasthttpproxy.FasthttpSocksDialer(t.Socks5)
		} else {
			c.Dial = fasthttpproxy.FasthttpProxyHTTPDialerTimeout(defaultHTTPTimeout)
		}
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

	var dftHeader fasthttp.RequestHeader
	dftHeader.SetMethod(fasthttp.MethodGet)
	dftHeader.SetContentType(defaultContentType)
	dftHeader.Set(fasthttp.HeaderConnection, "close")
	if t.KeepAlive {
		dftHeader.Set(fasthttp.HeaderConnection, "keep-alive")
	}

	if t.Token != "" {
		dftHeader.Set(fasthttp.HeaderAuthorization, t.Token)
	}

	if t.Header != nil {
		for k, v := range t.Header {
			dftHeader.Set(k, v)
		}
	}

	// 加载插件
	var client = &clientImpl{client: c, defaultHeader: &dftHeader}
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
