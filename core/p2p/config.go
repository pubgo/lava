package p2p

import (
	"time"

	"github.com/pubgo/lava/v2/core/p2p/turncred"
	"github.com/pubgo/lava/v2/core/tunnel"
)

// 默认 dev 服务器 coturn（见 deploy/coturn/turnserver.conf）。
const (
	DefaultSTUNURL      = "stun:118.178.168.253:3478"
	DefaultTURNURL      = "turn:118.178.168.253:3478"
	DefaultTURNUsername = "lava"
	DefaultTURNPassword = "lava-p2p-dev"
	DefaultTURNRealm    = "lava.dev"
)

// Config P2P 模块配置。
type Config struct {
	// STUNURLs STUN 服务器列表。
	STUNURLs []string `yaml:"stun_urls"`
	// TURN 配置（外部 coturn）。
	TURN TURNConfig `yaml:"turn"`
	// ICE 超时。
	ICETimeout time.Duration `yaml:"ice_timeout"`
	// SignalingAddr tunnel gateway 信令地址（P1 后接入 tunnel 控制流）。
	SignalingAddr string `yaml:"signaling_addr"`
	// AuthToken 注册/建连鉴权 token，对应 tunnel TokenAuthProvider。
	AuthToken string `yaml:"auth_token"`
	// Insecure 跳过 QUIC TLS 校验（仅开发）。
	Insecure bool `yaml:"insecure"`
	// TLS QUIC 证书（生产环境）；设置 cert/key 后自动启用 TLS。
	TLS TLSFileConfig `yaml:"tls"`
	// Reconnect 断线重连策略（Reconnect 使用；Dial 仍为单次尝试）。
	Reconnect ReconnectConfig `yaml:"reconnect"`
}

// ReconnectConfig 断线重连退避策略。
type ReconnectConfig struct {
	// MaxAttempts 最大尝试次数（含首次），默认 3。
	MaxAttempts int `yaml:"max_attempts"`
	// Backoff 重试间隔，默认 1s。
	Backoff time.Duration `yaml:"backoff"`
}

// TLSFileConfig QUIC TLS 文件路径。
type TLSFileConfig struct {
	CertFile string `yaml:"cert_file"`
	KeyFile  string `yaml:"key_file"`
	CAFile   string `yaml:"ca_file"`
}

// TURNConfig coturn 客户端配置。
type TURNConfig struct {
	URL      string `yaml:"url"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Realm    string `yaml:"realm"`
	// AuthSecret 与 coturn static-auth-secret 一致时，按 TURN REST API 生成临时凭证。
	AuthSecret string        `yaml:"auth_secret"`
	CredTTL    time.Duration `yaml:"cred_ttl"`
	// Disabled 为 true 时不使用 TURN（即使 URL 有默认值）。
	Disabled bool `yaml:"disabled"`
}

// Resolved 返回实际用于 ICE 的 TURN 凭证（含 HMAC 临时密码）。
func (t TURNConfig) Resolved(peerID string) (username, password string, err error) {
	if t.Disabled || t.URL == "" {
		return "", "", nil
	}
	if t.AuthSecret != "" {
		userID := t.Username
		if userID == "" {
			userID = peerID
		}
		return turncred.Generate(t.AuthSecret, userID, t.CredTTL, time.Time{})
	}
	return t.Username, t.Password, nil
}

// DefaultConfig 返回面向 dev 环境的默认配置。
func DefaultConfig() Config {
	return Config{
		STUNURLs: []string{DefaultSTUNURL},
		TURN: TURNConfig{
			URL:      DefaultTURNURL,
			Username: DefaultTURNUsername,
			Password: DefaultTURNPassword,
			Realm:    DefaultTURNRealm,
		},
		ICETimeout: 30 * time.Second,
		Reconnect: ReconnectConfig{
			MaxAttempts: 3,
			Backoff:     time.Second,
		},
	}
}

// ICEURLs 返回传给 pion/ice 的 STUN/TURN URL 列表（含 TURN 凭证）。
func (c Config) ICEURLs() []string {
	urls := append([]string{}, c.STUNURLs...)
	if c.TURN.URL != "" && !c.TURN.Disabled {
		urls = append(urls, c.TURN.URL)
	}
	return urls
}

// TransportOptions 转为 tunnel QUIC 传输选项。
func (c Config) TransportOptions() *tunnel.TransportOptions {
	opts := &tunnel.TransportOptions{
		Insecure:   c.Insecure,
		MaxStreams: 256,
		CertFile:   c.TLS.CertFile,
		KeyFile:    c.TLS.KeyFile,
		CAFile:     c.TLS.CAFile,
	}
	if opts.CertFile != "" && opts.KeyFile != "" {
		opts.EnableTLS = true
	}
	return opts
}
