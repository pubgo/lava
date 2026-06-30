// Package tunnelgateway 提供隧道代理网关实现
//
// 这个包实现了 tunnel.Gateway 接口，负责接收服务注册并暴露服务。
//
// 示例用法:
//
//	gw := tunnelgateway.New(&tunnelgateway.Config{
//		ListenAddr: ":9000",
//		Transport:  "yamux",
//		HTTPPort:   8080,
//		GRPCPort:   9090,
//		DebugPort:  6060,
//	})
//
//	if err := gw.Start(ctx); err != nil {
//		log.Fatal(err)
//	}
//	defer gw.Stop(ctx)
//
// 或使用 Builder 模式:
//
//	gw, err := tunnelgateway.NewBuilder().
//		WithListenAddr(":9000").
//		WithHTTPPort(8080).
//		WithDebugPort(6060).
//		Build()
package tunnelgateway

import (
	"context"
	"fmt"
	"net"

	"github.com/pubgo/lava/v2/core/tunnel"
)

// Config 是 Gateway 的配置
type Config = tunnel.GatewayConfig

// TransportOptions 传输层选项
type TransportOptions = tunnel.TransportOptions

// TLSConfig TLS 配置
type TLSConfig = tunnel.TLSConfig

// ServiceInfo 服务信息
type ServiceInfo = tunnel.ServiceInfo

// Endpoint 端点信息
type Endpoint = tunnel.Endpoint

// EndpointType 端点类型
type EndpointType = tunnel.EndpointType

// Status 网关状态
type Status = tunnel.GatewayStatus

// StatusInfo 网关状态信息
type StatusInfo = tunnel.GatewayStatusInfo

// 端点类型常量
const (
	EndpointTypeHTTP  = tunnel.EndpointTypeHTTP
	EndpointTypeGRPC  = tunnel.EndpointTypeGRPC
	EndpointTypeDebug = tunnel.EndpointTypeDebug
)

// 状态常量
const (
	StatusStopped  = tunnel.GatewayStatusStopped
	StatusStarting = tunnel.GatewayStatusStarting
	StatusRunning  = tunnel.GatewayStatusRunning
	StatusStopping = tunnel.GatewayStatusStopping
)

// 传输协议常量
const (
	TransportYamux = tunnel.TransportYamux
	TransportQUIC  = tunnel.TransportQUIC
	TransportKCP   = tunnel.TransportKCP
)

// Gateway 是隧道代理网关的封装
type Gateway struct {
	inner tunnel.Gateway
	cfg   *Config
}

// New 创建一个新的 Gateway 实例
func New(cfg *Config) *Gateway {
	return &Gateway{
		inner: NewGateway(cfg),
		cfg:   cfg,
	}
}

// Start 启动 Gateway，开始接受 Agent 连接
func (g *Gateway) Start(ctx context.Context) error {
	return g.inner.Start(ctx)
}

// Stop 停止 Gateway
func (g *Gateway) Stop(ctx context.Context) error {
	return g.inner.Stop(ctx)
}

// Services 获取所有已注册的服务
func (g *Gateway) Services() []*ServiceInfo {
	return g.inner.Services()
}

// GetService 获取指定名称的服务
func (g *Gateway) GetService(name string) (*ServiceInfo, error) {
	return g.inner.GetService(name)
}

// Status 获取 Gateway 当前状态
func (g *Gateway) Status() Status {
	return g.inner.Status()
}

// Forward 转发连接到指定服务的端点
func (g *Gateway) Forward(ctx context.Context, serviceName string, endpointType EndpointType, conn net.Conn) error {
	return g.inner.Forward(ctx, serviceName, endpointType, conn)
}

// Config 获取 Gateway 配置
func (g *Gateway) Config() *Config {
	return g.cfg
}

// Inner 获取底层的 tunnel.Gateway 实现
func (g *Gateway) Inner() tunnel.Gateway {
	return g.inner
}

// Builder Gateway 构建器
type Builder struct {
	cfg *Config
}

// NewBuilder 创建 Gateway 构建器
func NewBuilder() *Builder {
	cfg := tunnel.DefaultGatewayConfig()
	return &Builder{cfg: &cfg}
}

// WithListenAddr 设置监听地址
func (b *Builder) WithListenAddr(addr string) *Builder {
	b.cfg.ListenAddr = addr
	return b
}

// WithTransport 设置传输协议
func (b *Builder) WithTransport(transport string) *Builder {
	b.cfg.Transport = transport
	return b
}

// WithTransportOptions 设置传输层选项
func (b *Builder) WithTransportOptions(opts *TransportOptions) *Builder {
	b.cfg.TransportOptions = opts
	return b
}

// WithHTTPPort 设置 HTTP 端口
func (b *Builder) WithHTTPPort(port int) *Builder {
	b.cfg.HTTPPort = port
	return b
}

// WithGRPCPort 设置 gRPC 端口
func (b *Builder) WithGRPCPort(port int) *Builder {
	b.cfg.GRPCPort = port
	return b
}

// WithDebugPort 设置 Debug 端口
func (b *Builder) WithDebugPort(port int) *Builder {
	b.cfg.DebugPort = port
	return b
}

// WithHeartbeatInterval 设置心跳间隔（秒）
func (b *Builder) WithHeartbeatInterval(interval int) *Builder {
	b.cfg.HeartbeatInterval = interval
	return b
}

// WithHeartbeatTimeout 设置心跳超时（秒）
func (b *Builder) WithHeartbeatTimeout(timeout int) *Builder {
	b.cfg.HeartbeatTimeout = timeout
	return b
}

// WithHealthCheckInterval 设置健康检查间隔（秒）
func (b *Builder) WithHealthCheckInterval(interval int) *Builder {
	b.cfg.HealthCheckInterval = interval
	return b
}

// WithTLS 设置 TLS 配置
func (b *Builder) WithTLS(tls TLSConfig) *Builder {
	b.cfg.TLS = tls
	return b
}

// WithConfig 使用完整配置
func (b *Builder) WithConfig(cfg *Config) *Builder {
	b.cfg = cfg
	return b
}

// Build 构建 Gateway
func (b *Builder) Build() (*Gateway, error) {
	if b.cfg.ListenAddr == "" {
		return nil, fmt.Errorf("tunnelgateway: listen address is required")
	}
	if b.cfg.Transport == "" {
		b.cfg.Transport = TransportYamux
	}
	b.cfg.Normalize()
	return New(b.cfg), nil
}

// MustBuild 构建 Gateway，失败时 panic
func (b *Builder) MustBuild() *Gateway {
	gw, err := b.Build()
	if err != nil {
		panic(err)
	}
	return gw
}
