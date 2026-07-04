// Package p2p 提供基于 ICE/STUN/TURN 的点对点连接能力。
//
// ICE 打通 UDP 路径后在其上承载 QUIC（可靠 + 多路 + TLS 1.3），
// 信令复用 tunnel gateway 控制流；TURN 由外部 coturn 提供。
//
// 详见 docs/design-p2p.md。
package p2p

import (
	"context"
	"io"
	"net"
	"time"

	"github.com/pubgo/lava/v2/core/tunnel"
)

// Role 表示 ICE 协商中的角色。
type Role int

const (
	RoleDialer Role = iota
	RoleListener
)

// Coordinator 管理本节点的 P2P 连接（Listen / Dial / Reconnect）。
type Coordinator interface {
	Listen(ctx context.Context, selfID string) (Listener, error)
	Dial(ctx context.Context, peerID string) (PeerConn, error)
	// Reconnect 关闭与 peer 的现有连接并重新 ICE+QUIC 协商（带退避重试）。
	Reconnect(ctx context.Context, peerID string) (PeerConn, error)
	Close() error
	Stats() Stats
}

// Listener 接受来自对端的 P2P 入站连接。
type Listener interface {
	Accept(ctx context.Context) (PeerConn, error)
	Close() error
}

// PeerConn 是一条已建立的 P2P 连接，实现 tunnel.Session 语义。
type PeerConn interface {
	tunnel.Session
	RemotePeerID() string
	LocalPeerID() string
	// SelectedPair 返回 ICE 选路摘要（host/srflx/relay + 地址），用于可观测。
	SelectedPair() CandidatePairInfo
}

// CandidatePairInfo 是 ICE 选路结果摘要。
type CandidatePairInfo struct {
	LocalType  string `json:"local_type"`
	RemoteType string `json:"remote_type"`
	LocalAddr  string `json:"local_addr"`
	RemoteAddr string `json:"remote_addr"`
}

// Stats P2P 节点摘要。
type Stats struct {
	SelfID            string            `json:"self_id"`
	Listening         bool              `json:"listening"`
	ActiveConnections int               `json:"active_connections"`
	RelayConnections  int               `json:"relay_connections"`
	Connections       []ConnectionStats `json:"connections"`
}

// ConnectionStats 单条 P2P 连接的 ICE 选路摘要。
type ConnectionStats struct {
	RemotePeerID      string            `json:"remote_peer_id"`
	Pair              CandidatePairInfo `json:"pair"`
	ConnectedAt       time.Time         `json:"connected_at"`
	ConnectDurationMs int64             `json:"connect_duration_ms"`
	RTTMs             float64           `json:"rtt_ms,omitempty"`
	PairState         string            `json:"pair_state,omitempty"`
	Nominated         bool              `json:"nominated,omitempty"`
}

// PacketConn 是 ICE 打通后可用于 QUIC 的 UDP 数据报连接。
type PacketConn interface {
	net.PacketConn
	io.Closer
}
