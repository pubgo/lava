// Package tunnelagent 提供隧道代理客户端的便捷封装
//
// 这个包是 tunnel.Agent 的包装器，提供更简洁的 API 和使用体验。
//
// 示例用法:
//
//	agent := tunnelagent.New(&tunnelagent.Config{
//		GatewayAddr: "localhost:8080",
//		Transport:   "yamux",
//		ServiceName: "my-service",
//		Endpoints: []tunnelagent.EndpointConfig{
//			{Type: "http", LocalAddr: ":8081"},
//			{Type: "debug", LocalAddr: ":6060"},
//		},
//	})
//
//	if err := agent.Start(ctx); err != nil {
//		log.Fatal(err)
//	}
//	defer agent.Stop(ctx)
package tunnelagent

import (
	"context"

	"github.com/pubgo/lava/v2/core/tunnel"
)

// Config 是 Agent 的配置
type Config = tunnel.AgentConfig

// EndpointConfig 端点配置
type EndpointConfig = tunnel.EndpointConfig

// TransportOptions 传输层选项
type TransportOptions = tunnel.TransportOptions

// ServiceInfo 服务信息
type ServiceInfo = tunnel.ServiceInfo

// Endpoint 端点信息
type Endpoint = tunnel.Endpoint

// Status 代理状态
type Status = tunnel.AgentStatus

// Info Agent 信息
type Info = tunnel.AgentInfo

// 状态常量
const (
	StatusDisconnected = tunnel.StatusDisconnected
	StatusConnecting   = tunnel.StatusConnecting
	StatusConnected    = tunnel.StatusConnected
	StatusReconnecting = tunnel.StatusReconnecting
)

// Agent 是隧道代理客户端的封装
type Agent struct {
	inner tunnel.Agent
	cfg   *Config
}

// New 创建一个新的 Agent 实例
func New(cfg *Config) *Agent {
	return &Agent{
		inner: tunnel.NewAgent(cfg),
		cfg:   cfg,
	}
}

// Start 启动 Agent，连接到 Gateway
func (a *Agent) Start(ctx context.Context) error {
	return a.inner.Start(ctx)
}

// Stop 停止 Agent
func (a *Agent) Stop(ctx context.Context) error {
	return a.inner.Stop(ctx)
}

// Register 注册服务到 Gateway
func (a *Agent) Register(ctx context.Context, service *ServiceInfo) error {
	return a.inner.Register(ctx, service)
}

// Deregister 从 Gateway 注销服务
func (a *Agent) Deregister(ctx context.Context, serviceID string) error {
	return a.inner.Deregister(ctx, serviceID)
}

// Status 获取 Agent 当前状态
func (a *Agent) Status() Status {
	return a.inner.Status()
}

// Info 获取 Agent 信息
func (a *Agent) Info() *Info {
	return a.inner.Info()
}

// Config 获取 Agent 配置
func (a *Agent) Config() *Config {
	return a.cfg
}

// Inner 获取底层的 tunnel.Agent 实现
func (a *Agent) Inner() tunnel.Agent {
	return a.inner
}
