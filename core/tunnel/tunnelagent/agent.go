// Package tunnelagent 提供隧道代理客户端实现
//
// 这个包实现了 tunnel.Agent 接口，负责将本地服务注册到代理网关。
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
//
// 或使用 Builder 模式:
//
//	agent, err := tunnelagent.NewBuilder().
//		WithGatewayAddr("localhost:8080").
//		WithServiceName("my-service").
//		AddEndpoint("http", ":8081", "/").
//		AddEndpoint("debug", ":6060", "/debug").
//		Build()
package tunnelagent

import (
	"context"
	"fmt"

	"github.com/pubgo/lava/v2/core/tunnel"
)

// Config 是 Agent 的配置
type Config = tunnel.AgentConfig

// EndpointConfig 端点配置
type EndpointConfig = tunnel.EndpointConfig

// TransportOptions 传输层选项
type TransportOptions = tunnel.TransportOptions

// TLSConfig TLS 配置
type TLSConfig = tunnel.TLSConfig

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

// 传输协议常量
const (
	TransportYamux = tunnel.TransportYamux
	TransportQUIC  = tunnel.TransportQUIC
	TransportHTTP  = tunnel.TransportHTTP
	TransportKCP   = tunnel.TransportKCP
)

// Agent 是隧道代理客户端的封装
type Agent struct {
	inner tunnel.Agent
	cfg   *Config
}

// New 创建一个新的 Agent 实例
func New(cfg *Config) *Agent {
	return &Agent{
		inner: NewAgent(cfg),
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

// Builder Agent 构建器
type Builder struct {
	cfg *Config
}

// NewBuilder 创建 Agent 构建器
func NewBuilder() *Builder {
	cfg := tunnel.DefaultAgentConfig()
	return &Builder{cfg: &cfg}
}

// WithGatewayAddr 设置网关地址
func (b *Builder) WithGatewayAddr(addr string) *Builder {
	b.cfg.GatewayAddr = addr
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

// WithServiceID 设置服务 ID
func (b *Builder) WithServiceID(id string) *Builder {
	b.cfg.ServiceID = id
	return b
}

// WithServiceName 设置服务名称
func (b *Builder) WithServiceName(name string) *Builder {
	b.cfg.ServiceName = name
	return b
}

// WithServiceVersion 设置服务版本
func (b *Builder) WithServiceVersion(version string) *Builder {
	b.cfg.ServiceVersion = version
	return b
}

// WithMetadata 设置元数据
func (b *Builder) WithMetadata(metadata map[string]string) *Builder {
	b.cfg.Metadata = metadata
	return b
}

// WithEndpoints 设置端点配置
func (b *Builder) WithEndpoints(endpoints []EndpointConfig) *Builder {
	b.cfg.Endpoints = endpoints
	return b
}

// AddEndpoint 添加端点
func (b *Builder) AddEndpoint(endpointType, localAddr, path string) *Builder {
	b.cfg.Endpoints = append(b.cfg.Endpoints, EndpointConfig{
		Type:      endpointType,
		LocalAddr: localAddr,
		Path:      path,
	})
	return b
}

// WithHeartbeatInterval 设置心跳间隔（秒）
func (b *Builder) WithHeartbeatInterval(interval int) *Builder {
	b.cfg.HeartbeatInterval = interval
	return b
}

// WithReconnectInterval 设置重连间隔（秒）
func (b *Builder) WithReconnectInterval(interval int) *Builder {
	b.cfg.ReconnectInterval = interval
	return b
}

// WithMaxReconnectAttempts 设置最大重连次数
func (b *Builder) WithMaxReconnectAttempts(attempts int) *Builder {
	b.cfg.MaxReconnectAttempts = attempts
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

// Build 构建 Agent
func (b *Builder) Build() (*Agent, error) {
	if b.cfg.GatewayAddr == "" {
		return nil, fmt.Errorf("tunnelagent: gateway address is required")
	}
	if b.cfg.Transport == "" {
		b.cfg.Transport = TransportYamux
	}
	return New(b.cfg), nil
}

// MustBuild 构建 Agent，失败时 panic
func (b *Builder) MustBuild() *Agent {
	agent, err := b.Build()
	if err != nil {
		panic(err)
	}
	return agent
}
