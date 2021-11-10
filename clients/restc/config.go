package restc

import (
	"crypto/tls"
	"net/http"
	"runtime"
	"time"

	"github.com/pubgo/x/merge"
	"github.com/pubgo/xerror"

	"github.com/pubgo/lava/pkg/retry"
	"github.com/pubgo/lava/plugin"
	"github.com/pubgo/lava/types"
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
	backoff                       retry.Backoff
	tlsConfig                     *tls.Config
	Middlewares                   []string
}

func (t Cfg) Build(opts ...func(cfg *Cfg)) (_ Client, err error) {
	defer xerror.RespErr(&err)

	for i := range opts {
		opts[i](&t)
	}

	c := &http.Client{Transport: DefaultPooledTransport()}
	xerror.Panic(merge.CopyStruct(c, t))

	var certs []tls.Certificate
	if t.CertPath != "" && t.KeyPath != "" {
		c, err := tls.LoadX509KeyPair(t.CertPath, t.KeyPath)
		xerror.Panic(err)
		certs = append(certs, c)
	}
	//c.TLSConfig = &tls.Config{InsecureSkipVerify: t.Insecure, Certificates: certs}

	var middlewares []types.Middleware
	for _, name := range t.Middlewares {
		var mid = plugin.Get(name).Middleware()
		if mid == nil {
			continue
		}
		middlewares = append(middlewares, plugin.Get(name).Middleware())
	}

	// 加载插件
	var client = &clientImpl{client: c}
	client.do = doFunc(client)
	for i := len(middlewares); i > 0; i-- {
		client.do = middlewares[i-1](client.do)
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