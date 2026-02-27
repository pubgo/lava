package resty

import (
	"fmt"
	"net"
	"time"

	"github.com/pubgo/funk/v2/buildinfo/version"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttpproxy"
	"golang.org/x/net/http/httpproxy"
)

// Config 客户端配置结构
type Config struct {
	BaseUrl              string            `yaml:"base_url"`               // 基础 URL
	ServiceName          string            `yaml:"service_name"`           // 服务名称
	DefaultHeader        map[string]string `yaml:"default_header"`         // 默认请求头
	DefaultContentType   string            `yaml:"default_content_type"`   // 默认内容类型
	DefaultRetryCount    uint32            `yaml:"default_retry_count"`    // 默认重试次数
	DefaultRetryInterval time.Duration     `yaml:"default_retry_interval"` // 默认重试间隔
	BasicToken           string            `yaml:"basic_token"`            // Basic 认证令牌
	JwtToken             string            `yaml:"jwt_token"`              // JWT 认证令牌

	EnableProxy               bool          `yaml:"enable_proxy"`                 // 是否启用代理
	EnableAuth                bool          `yaml:"enable_auth"`                  // 是否启用认证
	DialTimeout               time.Duration `yaml:"dial_timeout"`                 // 拨号超时
	ReadTimeout               time.Duration `yaml:"read_timeout"`                 // 读取超时
	WriteTimeout              time.Duration `yaml:"write_timeout"`                // 写入超时
	MaxConnsPerHost           int           `yaml:"max_conns_per_host"`           // 每个主机的最大连接数
	MaxIdleConnDuration       time.Duration `yaml:"max_idle_conn_duration"`       // 最大空闲连接时长
	MaxIdemponentCallAttempts int           `yaml:"max_idemponent_call_attempts"` // 最大幂等调用尝试次数
	ReadBufferSize            int           `yaml:"read_buffer_size"`             // 读取缓冲区大小
	WriteBufferSize           int           `yaml:"write_buffer_size"`            // 写入缓冲区大小
	MaxResponseBodySize       int           `yaml:"max_response_body_size"`       // 最大响应体大小
}

// Build 构建 fasthttp 客户端
// 返回: 构建后的 fasthttp 客户端实例
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

// DefaultCfg 获取默认配置
// 返回: 默认配置实例
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
