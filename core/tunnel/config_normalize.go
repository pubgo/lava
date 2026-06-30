package tunnel

import (
	"os"
	"strings"
)

// MergeTransportOptions 将 TLS 配置合并进 TransportOptions，供传输层使用。
//
// 合并规则：
//   - 若 TransportOptions 为 nil，使用 DefaultTransportOptions()
//   - tls.enabled 或提供了 cert/key 文件时启用 EnableTLS
//   - TLS 块中的 cert/key/ca/insecure 仅在 TransportOptions 对应字段为空/false 时填充
func MergeTransportOptions(opts *TransportOptions, tls TLSConfig) *TransportOptions {
	if opts == nil {
		opts = DefaultTransportOptions()
	} else {
		merged := *opts
		opts = &merged
	}

	enableTLS := tls.Enabled || tls.CertFile != "" || tls.KeyFile != ""
	if enableTLS {
		opts.EnableTLS = true
	}
	if tls.CertFile != "" && opts.CertFile == "" {
		opts.CertFile = tls.CertFile
	}
	if tls.KeyFile != "" && opts.KeyFile == "" {
		opts.KeyFile = tls.KeyFile
	}
	if tls.CAFile != "" && opts.CAFile == "" {
		opts.CAFile = tls.CAFile
	}
	if tls.Insecure {
		opts.Insecure = true
	}

	return opts
}

// Normalize 合并 TLS 到 TransportOptions，应在创建 Gateway 前调用。
func (c *GatewayConfig) Normalize() {
	if c == nil {
		return
	}
	c.TransportOptions = MergeTransportOptions(c.TransportOptions, c.TLS)
}

// Normalize 合并 TLS 到 TransportOptions，应在创建 Agent 前调用。
func (c *AgentConfig) Normalize() {
	if c == nil {
		return
	}
	c.TransportOptions = MergeTransportOptions(c.TransportOptions, c.TLS)
}

// ApplyAuthTokenMetadata 将 auth_token 写入 Agent 注册元数据。
func ApplyAuthTokenMetadata(metadata map[string]string, token string) map[string]string {
	if token == "" {
		return metadata
	}
	if metadata == nil {
		metadata = make(map[string]string)
	}
	metadata["auth_token"] = token
	return metadata
}

// ConfigureGatewayAuth 为 Gateway 配置 token 鉴权，token 为空时不做变更。
func ConfigureGatewayAuth(gw Gateway, token string) error {
	if token == "" || gw == nil {
		return nil
	}
	provider, err := NewTokenAuthProvider(token)
	if err != nil {
		return err
	}
	gw.SetAuthProvider(provider)
	return nil
}

// ApplyTLSFromEnv 从环境变量填充 TLS 配置，未设置的环境变量会被忽略。
//
// 支持的环境变量（prefix 默认为 TUNNEL_TLS）：
//   - {prefix}_ENABLED=true
//   - {prefix}_CERT_FILE, {prefix}_KEY_FILE, {prefix}_CA_FILE
//   - {prefix}_INSECURE=true
func (t *TLSConfig) ApplyEnv(prefixes ...string) {
	if t == nil {
		return
	}
	prefix := "TUNNEL_TLS"
	if len(prefixes) > 0 && prefixes[0] != "" {
		prefix = prefixes[0]
	}

	if v := os.Getenv(prefix + "_ENABLED"); v == "1" || strings.EqualFold(v, "true") {
		t.Enabled = true
	}
	if v := os.Getenv(prefix + "_CERT_FILE"); v != "" {
		t.CertFile = v
	}
	if v := os.Getenv(prefix + "_KEY_FILE"); v != "" {
		t.KeyFile = v
	}
	if v := os.Getenv(prefix + "_CA_FILE"); v != "" {
		t.CAFile = v
	}
	if v := os.Getenv(prefix + "_INSECURE"); v == "1" || strings.EqualFold(v, "true") {
		t.Insecure = true
	}
}
