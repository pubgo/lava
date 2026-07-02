// Package tunnelagent 提供隧道代理客户端实现。
package tunnelagent

import (
	"context"
	"fmt"

	"github.com/pubgo/lava/v2/core/tunnel"
	_ "github.com/pubgo/lava/v2/core/tunnel/yamux" // 注册 yamux 传输
)

// StandaloneOptions 一次性启动 Agent 所需的参数。
type StandaloneOptions struct {
	// GatewayAddr gateway 地址（必填）。
	GatewayAddr string
	// ServiceName 注册到 gateway 的服务名（必填）。
	ServiceName string
	// ServiceVersion 服务版本，默认 "dev"。
	ServiceVersion string
	// AuthToken 注册鉴权 token，写入 metadata["auth_token"]。
	AuthToken string
	// Endpoints 要代理的本地端点。
	Endpoints []EndpointConfig
	// Transport 传输协议，默认 yamux。
	Transport string
	// Config 完整配置覆盖；非 nil 时在 env/字段默认值之上合并。
	Config *Config
}

// Standalone 一行启动 Agent：创建、连接 gateway、注册服务。
// 适合脚本与测试；生产常驻进程请用 CLI 或 DI 装配。
// 使用完毕调用 (*Agent).Stop。
func Standalone(ctx context.Context, opts StandaloneOptions) (*Agent, error) {
	if opts.GatewayAddr == "" {
		return nil, fmt.Errorf("tunnelagent: gateway addr required")
	}
	if opts.ServiceName == "" {
		return nil, fmt.Errorf("tunnelagent: service name required")
	}

	cfg := tunnel.AgentConfigFromEnv()
	cfg.GatewayAddr = opts.GatewayAddr
	cfg.ServiceName = opts.ServiceName
	if opts.ServiceVersion != "" {
		cfg.ServiceVersion = opts.ServiceVersion
	} else if cfg.ServiceVersion == "" {
		cfg.ServiceVersion = "dev"
	}
	if opts.Transport != "" {
		cfg.Transport = opts.Transport
	}
	if len(opts.Endpoints) > 0 {
		cfg.Endpoints = opts.Endpoints
	}
	if opts.AuthToken != "" {
		cfg.Metadata = tunnel.ApplyAuthTokenMetadata(cfg.Metadata, opts.AuthToken)
	} else if token := tunnel.AuthTokenFromEnv(); token != "" {
		cfg.Metadata = tunnel.ApplyAuthTokenMetadata(cfg.Metadata, token)
	}
	if opts.Config != nil {
		mergeAgentConfig(&cfg, opts.Config)
	}
	cfg.Normalize()

	agent := New(&cfg)
	if err := agent.Start(ctx); err != nil {
		return nil, err
	}
	return agent, nil
}

// mergeAgentConfig 用 override 中非零字段覆盖 base。
func mergeAgentConfig(base, override *Config) {
	if override.GatewayAddr != "" {
		base.GatewayAddr = override.GatewayAddr
	}
	if override.Transport != "" {
		base.Transport = override.Transport
	}
	if override.ServiceID != "" {
		base.ServiceID = override.ServiceID
	}
	if override.ServiceName != "" {
		base.ServiceName = override.ServiceName
	}
	if override.ServiceVersion != "" {
		base.ServiceVersion = override.ServiceVersion
	}
	if override.Metadata != nil {
		base.Metadata = override.Metadata
	}
	if len(override.Endpoints) > 0 {
		base.Endpoints = override.Endpoints
	}
	if override.HeartbeatInterval > 0 {
		base.HeartbeatInterval = override.HeartbeatInterval
	}
	if override.ReconnectInterval > 0 {
		base.ReconnectInterval = override.ReconnectInterval
	}
	if override.MaxReconnectAttempts > 0 {
		base.MaxReconnectAttempts = override.MaxReconnectAttempts
	}
	if override.TransportOptions != nil {
		base.TransportOptions = override.TransportOptions
	}
	if override.TLS.Enabled || override.TLS.CertFile != "" || override.TLS.KeyFile != "" {
		base.TLS = override.TLS
	}
	if override.P2PSignalHandler != nil {
		base.P2PSignalHandler = override.P2PSignalHandler
	}
}
