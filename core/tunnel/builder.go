package tunnel

import (
	"context"
	"fmt"

	"github.com/pubgo/funk/v2/log"
)

// AgentBuilder Agent 构建器
type AgentBuilder struct {
	cfg *AgentConfig
}

// NewAgentBuilder 创建 Agent 构建器
func NewAgentBuilder() *AgentBuilder {
	cfg := DefaultAgentConfig()
	return &AgentBuilder{cfg: &cfg}
}

// WithGatewayAddr 设置网关地址
func (b *AgentBuilder) WithGatewayAddr(addr string) *AgentBuilder {
	b.cfg.GatewayAddr = addr
	return b
}

// WithTransport 设置传输协议
func (b *AgentBuilder) WithTransport(transport string) *AgentBuilder {
	b.cfg.Transport = transport
	return b
}

// WithTransportOptions 设置传输层选项
func (b *AgentBuilder) WithTransportOptions(opts *TransportOptions) *AgentBuilder {
	b.cfg.TransportOptions = opts
	return b
}

// WithServiceID 设置服务 ID
func (b *AgentBuilder) WithServiceID(id string) *AgentBuilder {
	b.cfg.ServiceID = id
	return b
}

// WithServiceName 设置服务名称
func (b *AgentBuilder) WithServiceName(name string) *AgentBuilder {
	b.cfg.ServiceName = name
	return b
}

// WithServiceVersion 设置服务版本
func (b *AgentBuilder) WithServiceVersion(version string) *AgentBuilder {
	b.cfg.ServiceVersion = version
	return b
}

// WithMetadata 设置元数据
func (b *AgentBuilder) WithMetadata(metadata map[string]string) *AgentBuilder {
	b.cfg.Metadata = metadata
	return b
}

// WithEndpoints 设置端点配置
func (b *AgentBuilder) WithEndpoints(endpoints []EndpointConfig) *AgentBuilder {
	b.cfg.Endpoints = endpoints
	return b
}

// AddEndpoint 添加端点
func (b *AgentBuilder) AddEndpoint(endpointType, localAddr, path string) *AgentBuilder {
	b.cfg.Endpoints = append(b.cfg.Endpoints, EndpointConfig{
		Type:      endpointType,
		LocalAddr: localAddr,
		Path:      path,
	})
	return b
}

// WithHeartbeatInterval 设置心跳间隔（秒）
func (b *AgentBuilder) WithHeartbeatInterval(interval int) *AgentBuilder {
	b.cfg.HeartbeatInterval = interval
	return b
}

// WithReconnectInterval 设置重连间隔（秒）
func (b *AgentBuilder) WithReconnectInterval(interval int) *AgentBuilder {
	b.cfg.ReconnectInterval = interval
	return b
}

// WithMaxReconnectAttempts 设置最大重连次数
func (b *AgentBuilder) WithMaxReconnectAttempts(attempts int) *AgentBuilder {
	b.cfg.MaxReconnectAttempts = attempts
	return b
}

// WithTLS 设置 TLS 配置
func (b *AgentBuilder) WithTLS(tls TLSConfig) *AgentBuilder {
	b.cfg.TLS = tls
	return b
}

// WithConfig 使用完整配置
func (b *AgentBuilder) WithConfig(cfg *AgentConfig) *AgentBuilder {
	b.cfg = cfg
	return b
}

// Build 构建 Agent
func (b *AgentBuilder) Build() (Agent, error) {
	if b.cfg.GatewayAddr == "" {
		return nil, fmt.Errorf("tunnel: gateway address is required")
	}
	if b.cfg.Transport == "" {
		b.cfg.Transport = TransportYamux
	}
	return NewAgent(b.cfg), nil
}

// MustBuild 构建 Agent，失败时 panic
func (b *AgentBuilder) MustBuild() Agent {
	agent, err := b.Build()
	if err != nil {
		panic(err)
	}
	return agent
}

// GatewayBuilder Gateway 构建器
type GatewayBuilder struct {
	cfg *GatewayConfig
}

// NewGatewayBuilder 创建 Gateway 构建器
func NewGatewayBuilder() *GatewayBuilder {
	cfg := DefaultGatewayConfig()
	return &GatewayBuilder{cfg: &cfg}
}

// WithListenAddr 设置监听地址
func (b *GatewayBuilder) WithListenAddr(addr string) *GatewayBuilder {
	b.cfg.ListenAddr = addr
	return b
}

// WithTransport 设置传输协议
func (b *GatewayBuilder) WithTransport(transport string) *GatewayBuilder {
	b.cfg.Transport = transport
	return b
}

// WithTransportOptions 设置传输层选项
func (b *GatewayBuilder) WithTransportOptions(opts *TransportOptions) *GatewayBuilder {
	b.cfg.TransportOptions = opts
	return b
}

// WithHTTPPort 设置 HTTP 端口
func (b *GatewayBuilder) WithHTTPPort(port int) *GatewayBuilder {
	b.cfg.HTTPPort = port
	return b
}

// WithGRPCPort 设置 gRPC 端口
func (b *GatewayBuilder) WithGRPCPort(port int) *GatewayBuilder {
	b.cfg.GRPCPort = port
	return b
}

// WithDebugPort 设置 Debug 端口
func (b *GatewayBuilder) WithDebugPort(port int) *GatewayBuilder {
	b.cfg.DebugPort = port
	return b
}

// WithHeartbeatInterval 设置心跳间隔（秒）
func (b *GatewayBuilder) WithHeartbeatInterval(interval int) *GatewayBuilder {
	b.cfg.HeartbeatInterval = interval
	return b
}

// WithHeartbeatTimeout 设置心跳超时（秒）
func (b *GatewayBuilder) WithHeartbeatTimeout(timeout int) *GatewayBuilder {
	b.cfg.HeartbeatTimeout = timeout
	return b
}

// WithHealthCheckInterval 设置健康检查间隔（秒）
func (b *GatewayBuilder) WithHealthCheckInterval(interval int) *GatewayBuilder {
	b.cfg.HealthCheckInterval = interval
	return b
}

// WithTLS 设置 TLS 配置
func (b *GatewayBuilder) WithTLS(tls TLSConfig) *GatewayBuilder {
	b.cfg.TLS = tls
	return b
}

// WithConfig 使用完整配置
func (b *GatewayBuilder) WithConfig(cfg *GatewayConfig) *GatewayBuilder {
	b.cfg = cfg
	return b
}

// Build 构建 Gateway
func (b *GatewayBuilder) Build() (Gateway, error) {
	if b.cfg.ListenAddr == "" {
		return nil, fmt.Errorf("tunnel: listen address is required")
	}
	if b.cfg.Transport == "" {
		b.cfg.Transport = TransportYamux
	}
	return NewGateway(b.cfg), nil
}

// MustBuild 构建 Gateway，失败时 panic
func (b *GatewayBuilder) MustBuild() Gateway {
	gw, err := b.Build()
	if err != nil {
		panic(err)
	}
	return gw
}

// StartAgent 便捷函数：创建并启动 Agent
func StartAgent(ctx context.Context, gatewayAddr string, services ...*ServiceInfo) (Agent, error) {
	agent, err := NewAgentBuilder().
		WithGatewayAddr(gatewayAddr).
		Build()
	if err != nil {
		return nil, err
	}

	if err := agent.Start(ctx); err != nil {
		return nil, err
	}

	for _, svc := range services {
		if err := agent.Register(ctx, svc); err != nil {
			log.Warn().Err(err).Str("service", svc.Name).Msg("Failed to register service")
		}
	}

	return agent, nil
}

// StartGateway 便捷函数：创建并启动 Gateway
func StartGateway(ctx context.Context, listenAddr string) (Gateway, error) {
	gw, err := NewGatewayBuilder().
		WithListenAddr(listenAddr).
		Build()
	if err != nil {
		return nil, err
	}

	if err := gw.Start(ctx); err != nil {
		return nil, err
	}

	return gw, nil
}
