package p2p

import (
	"context"
	"net"
	"net/http"
	"sync"
)

// Client 是 Coordinator 之上的高层封装：按 peerID 复用 P2P 连接、
// 在连接失效时自动重连，并提供 net.Conn / gRPC / HTTP 适配，便于集成。
//
// 典型用法：
//
//	client := p2p.NewClient(coord)
//	conn, err := client.OpenStream(ctx, "node-b") // 返回 net.Conn
//
//	// gRPC over P2P：
//	cc, _ := grpc.NewClient("passthrough:///node-b",
//	    grpc.WithContextDialer(client.NetDialer()),
//	    grpc.WithTransportCredentials(insecure.NewCredentials()))
//
//	// HTTP over P2P（host 即 peerID）：
//	resp, _ := client.HTTPClient().Get("http://node-b/healthz")
type Client struct {
	coord Coordinator

	mu    sync.Mutex
	conns map[string]*clientEntry
}

type clientEntry struct {
	mu sync.Mutex
	pc PeerConn
}

// NewClient 基于已有 Coordinator 创建池化客户端。
func NewClient(coord Coordinator) *Client {
	return &Client{
		coord: coord,
		conns: make(map[string]*clientEntry),
	}
}

func (c *Client) entry(peerID string) *clientEntry {
	c.mu.Lock()
	defer c.mu.Unlock()
	e := c.conns[peerID]
	if e == nil {
		e = &clientEntry{}
		c.conns[peerID] = e
	}
	return e
}

// Conn 返回到 peerID 的一条健康连接：命中缓存则复用，
// 连接已关闭则自动重连，无连接则首次拨号。并发调用同一 peer 会串行化。
func (c *Client) Conn(ctx context.Context, peerID string) (PeerConn, error) {
	if peerID == "" {
		return nil, ErrPeerNotFound
	}
	e := c.entry(peerID)
	e.mu.Lock()
	defer e.mu.Unlock()
	return c.connLocked(ctx, peerID, e)
}

func (c *Client) connLocked(ctx context.Context, peerID string, e *clientEntry) (PeerConn, error) {
	if e.pc != nil && !e.pc.IsClosed() {
		return e.pc, nil
	}
	var (
		pc  PeerConn
		err error
	)
	if e.pc != nil {
		pc, err = c.coord.Reconnect(ctx, peerID)
	} else {
		pc, err = c.coord.Dial(ctx, peerID)
	}
	if err != nil {
		return nil, err
	}
	e.pc = pc
	return pc, nil
}

// OpenStream 在到 peerID 的复用连接上打开一条新流（即 net.Conn）。
// 若底层连接在打开时已失效，会重连一次后重试。
func (c *Client) OpenStream(ctx context.Context, peerID string) (net.Conn, error) {
	e := c.entry(peerID)
	e.mu.Lock()
	defer e.mu.Unlock()

	pc, err := c.connLocked(ctx, peerID, e)
	if err != nil {
		return nil, err
	}
	st, err := pc.Open(ctx)
	if err != nil {
		// 连接可能在 check 与 open 之间失效：强制重连一次再试。
		pc, rerr := c.coord.Reconnect(ctx, peerID)
		if rerr != nil {
			return nil, err
		}
		e.pc = pc
		st, err = pc.Open(ctx)
		if err != nil {
			return nil, err
		}
	}
	return st, nil
}

// NetDialer 返回可用于 grpc.WithContextDialer 的拨号函数（addr 即 peerID）。
func (c *Client) NetDialer() func(context.Context, string) (net.Conn, error) {
	return func(ctx context.Context, addr string) (net.Conn, error) {
		return c.OpenStream(ctx, peerIDFromAddr(addr))
	}
}

// HTTPTransport 返回经 P2P 拨号的 http.Transport（URL host 即 peerID）。
func (c *Client) HTTPTransport() *http.Transport {
	return &http.Transport{
		DialContext: func(ctx context.Context, _ string, addr string) (net.Conn, error) {
			return c.OpenStream(ctx, peerIDFromAddr(addr))
		},
		// P2P 已是单条 QUIC 多路复用，禁用 http 自带的连接池压缩握手抖动。
		DisableKeepAlives: false,
	}
}

// HTTPClient 返回经 P2P 拨号的 http.Client（URL host 即 peerID）。
func (c *Client) HTTPClient() *http.Client {
	return &http.Client{Transport: c.HTTPTransport()}
}

// Close 关闭所有池化连接（不关闭底层 Coordinator）。
func (c *Client) Close() error {
	c.mu.Lock()
	entries := make([]*clientEntry, 0, len(c.conns))
	for _, e := range c.conns {
		entries = append(entries, e)
	}
	c.conns = make(map[string]*clientEntry)
	c.mu.Unlock()

	for _, e := range entries {
		e.mu.Lock()
		if e.pc != nil {
			_ = e.pc.Close()
			e.pc = nil
		}
		e.mu.Unlock()
	}
	return nil
}

// peerIDFromAddr 从 grpc/http 传入的 addr 中提取 peerID（去掉可能的 :port）。
func peerIDFromAddr(addr string) string {
	if host, _, err := net.SplitHostPort(addr); err == nil {
		return host
	}
	return addr
}
