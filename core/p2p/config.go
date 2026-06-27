package p2p

import (
	"time"

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
}

// TURNConfig coturn 客户端配置。
type TURNConfig struct {
	URL      string `yaml:"url"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Realm    string `yaml:"realm"`
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
	}
}

// ICEURLs 返回传给 pion/ice 的 STUN/TURN URL 列表（含 TURN 凭证）。
func (c Config) ICEURLs() []string {
	urls := append([]string{}, c.STUNURLs...)
	if c.TURN.URL != "" {
		urls = append(urls, c.TURN.URL)
	}
	return urls
}

// TransportOptions 转为 tunnel QUIC 传输选项。
func (c Config) TransportOptions() *tunnel.TransportOptions {
	return &tunnel.TransportOptions{
		Insecure:   c.Insecure,
		MaxStreams: 256,
	}
}
