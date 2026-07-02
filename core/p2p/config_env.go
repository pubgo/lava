package p2p

import (
	"fmt"
	"os"
	"strings"
	"time"
)

// ConfigFromEnv 从环境变量加载 P2P 配置（在 DefaultConfig 基础上覆盖）。
//
// 环境变量：
//   - P2P_STUN_URLS       逗号分隔 STUN URL
//   - P2P_TURN_URL        TURN URL（设为空或 P2P_TURN_DISABLED=1 禁用）
//   - P2P_TURN_DISABLED   true/1 禁用 TURN
//   - P2P_TURN_USER       TURN 用户名（HMAC 模式下作为 user_id 后缀，缺省用 peerID）
//   - P2P_TURN_PASS       TURN 静态密码（未设置 P2P_TURN_SECRET 时使用）
//   - P2P_TURN_SECRET     coturn static-auth-secret，启用 HMAC 临时凭证
//   - P2P_TURN_CRED_TTL   临时凭证 TTL（如 24h，默认 24h）
//   - P2P_RECONNECT_MAX_ATTEMPTS  Reconnect 最大尝试次数（默认 3）
//   - P2P_RECONNECT_BACKOFF       Reconnect 重试间隔（如 1s）
//   - P2P_INSECURE        true/1 跳过 QUIC TLS 校验（仅开发）
//   - P2P_CERT_FILE       QUIC TLS 证书
//   - P2P_KEY_FILE        QUIC TLS 私钥
//   - P2P_CA_FILE         QUIC TLS CA（可选）
//   - TUNNEL_AUTH_TOKEN   信令鉴权 token（若 P2P_AUTH_TOKEN 未设置）
//   - P2P_AUTH_TOKEN      信令鉴权 token
func ConfigFromEnv() Config {
	cfg := DefaultConfig()
	if v := strings.TrimSpace(os.Getenv("P2P_STUN_URLS")); v != "" {
		cfg.STUNURLs = splitComma(v)
	}
	if v := os.Getenv("P2P_TURN_URL"); v != "" {
		cfg.TURN.URL = v
	}
	if v := os.Getenv("P2P_TURN_DISABLED"); v == "1" || strings.EqualFold(v, "true") {
		cfg.TURN.Disabled = true
	}
	if v := os.Getenv("P2P_TURN_USER"); v != "" {
		cfg.TURN.Username = v
	}
	if v := os.Getenv("P2P_TURN_PASS"); v != "" {
		cfg.TURN.Password = v
	}
	if v := os.Getenv("P2P_TURN_SECRET"); v != "" {
		cfg.TURN.AuthSecret = v
	}
	if v := os.Getenv("P2P_TURN_CRED_TTL"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			cfg.TURN.CredTTL = d
		}
	}
	if v := os.Getenv("P2P_ICE_TIMEOUT"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			cfg.ICETimeout = d
		}
	}
	if v := os.Getenv("P2P_RECONNECT_MAX_ATTEMPTS"); v != "" {
		if n, err := parseIntEnv(v); err == nil && n > 0 {
			cfg.Reconnect.MaxAttempts = n
		}
	}
	if v := os.Getenv("P2P_RECONNECT_BACKOFF"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			cfg.Reconnect.Backoff = d
		}
	}
	if v := os.Getenv("P2P_INSECURE"); v == "1" || strings.EqualFold(v, "true") {
		cfg.Insecure = true
	}
	if v := os.Getenv("P2P_CERT_FILE"); v != "" {
		cfg.TLS.CertFile = v
	}
	if v := os.Getenv("P2P_KEY_FILE"); v != "" {
		cfg.TLS.KeyFile = v
	}
	if v := os.Getenv("P2P_CA_FILE"); v != "" {
		cfg.TLS.CAFile = v
	}
	if token := AuthTokenFromEnv(); token != "" {
		cfg.AuthToken = token
	}
	return cfg
}

// PeerIDFromEnv 读取 P2P 节点 ID；未配置时返回空字符串（表示不启用 P2P）。
func PeerIDFromEnv() string {
	return strings.TrimSpace(os.Getenv("P2P_PEER_ID"))
}

// AuthTokenFromEnv 读取 P2P 信令鉴权 token。
func AuthTokenFromEnv() string {
	if v := os.Getenv("P2P_AUTH_TOKEN"); v != "" {
		return v
	}
	return os.Getenv("TUNNEL_AUTH_TOKEN")
}

func splitComma(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}

func parseIntEnv(v string) (int, error) {
	var n int
	_, err := fmt.Sscanf(strings.TrimSpace(v), "%d", &n)
	return n, err
}
