package p2p

import (
	"os"
	"strings"
	"time"
)

// ConfigFromEnv 从环境变量加载 P2P 配置（在 DefaultConfig 基础上覆盖）。
//
// 环境变量：
//   - P2P_STUN_URLS       逗号分隔 STUN URL
//   - P2P_TURN_URL        TURN URL
//   - P2P_TURN_USER       TURN 用户名
//   - P2P_TURN_PASS       TURN 密码
//   - P2P_ICE_TIMEOUT     ICE 超时（如 30s）
//   - P2P_INSECURE        true/1 跳过 QUIC TLS 校验（仅开发）
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
	if v := os.Getenv("P2P_TURN_USER"); v != "" {
		cfg.TURN.Username = v
	}
	if v := os.Getenv("P2P_TURN_PASS"); v != "" {
		cfg.TURN.Password = v
	}
	if v := os.Getenv("P2P_ICE_TIMEOUT"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			cfg.ICETimeout = d
		}
	}
	if v := os.Getenv("P2P_INSECURE"); v == "1" || strings.EqualFold(v, "true") {
		cfg.Insecure = true
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
