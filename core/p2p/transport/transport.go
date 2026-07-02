package transport

import (
	"context"
	"net"
	"sync"

	"github.com/pubgo/lava/v2/core/p2p"
	"github.com/pubgo/lava/v2/core/tunnel"
)

func init() {
	tunnel.RegisterTransport(TransportName, NewTransport)
}

// TransportName 与 tunnel.TransportP2P 相同。
const TransportName = tunnel.TransportP2P

var (
	mu          sync.RWMutex
	coordinator p2p.Coordinator
)

// SetCoordinator 注册全局 P2P 协调器（Dial/Listen 前必须设置）。
func SetCoordinator(c p2p.Coordinator) {
	mu.Lock()
	defer mu.Unlock()
	coordinator = c
}

// ClearCoordinator 注销协调器；仅当当前注册实例与 c 相同时才清除（c 为 nil 时强制清除）。
func ClearCoordinator(c p2p.Coordinator) {
	mu.Lock()
	defer mu.Unlock()
	if c == nil || coordinator == c {
		coordinator = nil
	}
}

// NewTransport 创建 P2P tunnel 传输（addr 语义为 peerID）。
func NewTransport(opts *tunnel.TransportOptions) (tunnel.Transport, error) {
	return &p2pTransport{opts: opts}, nil
}

type p2pTransport struct {
	opts *tunnel.TransportOptions
}

func (t *p2pTransport) Name() string { return TransportName }

func (t *p2pTransport) Dial(ctx context.Context, peerID string) (tunnel.Session, error) {
	c := getCoordinator()
	if c == nil {
		return nil, p2p.ErrPeerNotFound
	}
	pc, err := c.Dial(ctx, peerID)
	if err != nil {
		return nil, err
	}
	return pc, nil
}

func (t *p2pTransport) Listen(ctx context.Context, selfID string) (tunnel.Listener, error) {
	c := getCoordinator()
	if c == nil {
		return nil, p2p.ErrSignalingClosed
	}
	ln, err := c.Listen(ctx, selfID)
	if err != nil {
		return nil, err
	}
	return &p2pListener{inner: ln}, nil
}

func getCoordinator() p2p.Coordinator {
	mu.RLock()
	defer mu.RUnlock()
	return coordinator
}

type p2pListener struct {
	inner p2p.Listener
}

func (l *p2pListener) Accept() (tunnel.Session, error) {
	pc, err := l.inner.Accept(context.Background())
	if err != nil {
		return nil, err
	}
	return pc, nil
}

func (l *p2pListener) Close() error { return l.inner.Close() }
func (l *p2pListener) Addr() net.Addr {
	return &peerAddr{}
}

type peerAddr struct{}

func (peerAddr) Network() string { return "p2p" }
func (peerAddr) String() string  { return "p2p" }
