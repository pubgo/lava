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
	Trace                         bool
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
	Middlewares                   []types.Middleware
	BasePath                      string
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
		_c, _err := tls.LoadX509KeyPair(t.CertPath, t.KeyPath)
		xerror.Panic(_err)
		certs = append(certs, _c)
	}
	t.tlsConfig = &tls.Config{InsecureSkipVerify: t.Insecure, Certificates: certs}

	var middlewares []types.Middleware

	// 加载插件
	// 加载全局
	for _, plg := range plugin.All() {
		if plg.Middleware() == nil {
			continue
		}
		middlewares = append(middlewares, plg.Middleware())
	}

	// 最后加载业务自定义
	middlewares = append(middlewares, t.Middlewares...)

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
		// http client缓存数量
		MaxConnsPerHost: runtime.GOMAXPROCS(0) + 1,
	}
}
