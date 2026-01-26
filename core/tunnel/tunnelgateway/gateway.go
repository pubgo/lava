// Package tunnelgateway 提供隧道代理网关的便捷封装
//
// 这个包是 tunnel.Gateway 的包装器，提供更简洁的 API 和使用体验。
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
package tunnelgateway

import (
	"context"
	"net"

	"github.com/pubgo/lava/v2/core/tunnel"
)

// Config 是 Gateway 的配置
type Config = tunnel.GatewayConfig

// TransportOptions 传输层选项
type TransportOptions = tunnel.TransportOptions

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
	StatusRunning  = tunnel.GatewayStatusRunning
	StatusStopping = tunnel.GatewayStatusStopping
)

// Gateway 是隧道代理网关的封装
type Gateway struct {
	inner tunnel.Gateway
	cfg   *Config
}

// New 创建一个新的 Gateway 实例
func New(cfg *Config) *Gateway {
	return &Gateway{
		inner: tunnel.NewGateway(cfg),
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
