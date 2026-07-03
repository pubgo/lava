// 环境变量加载：与 CLI（cmds/tunnelcmd）使用同一套 TUNNEL_* 命名。
package tunnel

import (
	"fmt"
	"os"
	"strings"
)

// AuthTokenFromEnv 读取 tunnel 鉴权 token（Agent 注册与 Gateway 校验共用）。
func AuthTokenFromEnv() string {
	return strings.TrimSpace(os.Getenv("TUNNEL_AUTH_TOKEN"))
}

// AdminTokenFromEnv 读取 Admin UI 鉴权 token（TUNNEL_ADMIN_TOKEN）。
func AdminTokenFromEnv() string {
	return strings.TrimSpace(os.Getenv("TUNNEL_ADMIN_TOKEN"))
}

// GatewayConfigFromEnv 从环境变量加载 Gateway 配置（在 DefaultGatewayConfig 基础上覆盖）。
//
// 环境变量：
//   - TUNNEL_LISTEN_ADDR          Agent 连接监听地址（默认 :7007）
//   - TUNNEL_HTTP_PORT            HTTP 代理端口（默认 8080）
//   - TUNNEL_GRPC_PORT            gRPC 代理端口（默认 9090）
//   - TUNNEL_DEBUG_PORT           Debug 代理端口（默认 6060）
//   - TUNNEL_P2P_SIGNAL_RATE_LIMIT   每 peer 每秒 P2P 信令条数（默认 60）
//   - TUNNEL_P2P_REGISTER_RATE_LIMIT 每 agent 每秒 P2P 注册次数（默认 10）
//   - TUNNEL_TLS_*                见 TLSConfig.ApplyEnv
func GatewayConfigFromEnv() GatewayConfig {
	cfg := DefaultGatewayConfig()
	if v := strings.TrimSpace(os.Getenv("TUNNEL_LISTEN_ADDR")); v != "" {
		cfg.ListenAddr = v
	}
	if v := os.Getenv("TUNNEL_HTTP_PORT"); v != "" {
		if p, err := parsePortEnv(v); err == nil {
			cfg.HTTPPort = p
		}
	}
	if v := os.Getenv("TUNNEL_GRPC_PORT"); v != "" {
		if p, err := parsePortEnv(v); err == nil {
			cfg.GRPCPort = p
		}
	}
	if v := os.Getenv("TUNNEL_DEBUG_PORT"); v != "" {
		if p, err := parsePortEnv(v); err == nil {
			cfg.DebugPort = p
		}
	}
	if v := os.Getenv("TUNNEL_P2P_SIGNAL_RATE_LIMIT"); v != "" {
		if n, err := parseIntEnv(v); err == nil && n > 0 {
			cfg.P2PSignalRateLimit = n
		}
	}
	if v := os.Getenv("TUNNEL_P2P_REGISTER_RATE_LIMIT"); v != "" {
		if n, err := parseIntEnv(v); err == nil && n > 0 {
			cfg.P2PRegisterRateLimit = n
		}
	}
	cfg.TLS.ApplyEnv()
	return cfg
}

// AgentConfigFromEnv 从环境变量加载 Agent 配置（在 DefaultAgentConfig 基础上覆盖）。
//
// 环境变量：
//   - TUNNEL_GATEWAY_ADDR         Gateway 地址（必填方可连接）
//   - SERVICE_NAME                服务名（CLI 默认用项目名）
//   - TUNNEL_AUTH_TOKEN           写入 metadata["auth_token"]
//   - TUNNEL_TLS_*                见 TLSConfig.ApplyEnv
func AgentConfigFromEnv() AgentConfig {
	cfg := DefaultAgentConfig()
	if v := strings.TrimSpace(os.Getenv("TUNNEL_GATEWAY_ADDR")); v != "" {
		cfg.GatewayAddr = v
	}
	if v := strings.TrimSpace(os.Getenv("SERVICE_NAME")); v != "" {
		cfg.ServiceName = v
	}
	if token := AuthTokenFromEnv(); token != "" {
		cfg.Metadata = ApplyAuthTokenMetadata(cfg.Metadata, token)
	}
	cfg.TLS.ApplyEnv()
	return cfg
}

func parsePortEnv(v string) (int, error) {
	var port int
	_, err := fmt.Sscanf(strings.TrimSpace(v), "%d", &port)
	return port, err
}

func parseIntEnv(v string) (int, error) {
	var n int
	_, err := fmt.Sscanf(strings.TrimSpace(v), "%d", &n)
	return n, err
}
