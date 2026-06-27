package p2p

import (
	"context"
	"io"
	"net"
	"sync"

	pionice "github.com/pion/ice/v4"

	"github.com/pubgo/lava/v2/core/p2p/ice"
	"github.com/pubgo/lava/v2/core/p2p/quicconn"
	"github.com/pubgo/lava/v2/core/p2p/signaling"
	"github.com/pubgo/lava/v2/core/tunnel"
)

type coordinator struct {
	cfg      Config
	broker   signaling.Broker
	selfID   string
	transOpt *tunnel.TransportOptions

	mu        sync.Mutex
	closed    bool
	inbound   chan *peerConn
	stopCh    chan struct{}
	listener  *tunnelListener
	connMu    sync.Mutex
	active    map[string]ConnectionStats
}

// NewCoordinator 创建 P2P 协调器。
func NewCoordinator(cfg Config, broker signaling.Broker, selfID string) Coordinator {
	if cfg.ICETimeout <= 0 {
		cfg.ICETimeout = DefaultConfig().ICETimeout
	}
	return &coordinator{
		cfg:      cfg,
		broker:   broker,
		selfID:   selfID,
		transOpt: cfg.TransportOptions(),
		inbound:  make(chan *peerConn),
		stopCh:   make(chan struct{}),
		active:   make(map[string]ConnectionStats),
	}
}

func (c *coordinator) Listen(ctx context.Context, selfID string) (Listener, error) {
	if selfID != "" && selfID != c.selfID {
		return nil, ErrPeerNotFound
	}
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return nil, ErrSignalingClosed
	}
	if c.listener != nil {
		c.mu.Unlock()
		return c.listener, nil
	}
	ln := &tunnelListener{coord: c, ctx: ctx}
	c.listener = ln
	c.mu.Unlock()

	go c.acceptLoop(ctx)
	return ln, nil
}

func (c *coordinator) Dial(ctx context.Context, peerID string) (PeerConn, error) {
	if peerID == "" {
		return nil, ErrPeerNotFound
	}
	iceCtx, cancel := context.WithTimeout(ctx, c.cfg.ICETimeout)
	defer cancel()

	iceRes, err := ice.Connect(iceCtx, toICEConfig(c.cfg), c.broker, c.selfID, peerID, ice.RoleDialer)
	if err != nil {
		return nil, err
	}
	pc, err := c.quicDial(iceCtx, iceRes)
	return c.attachConn(pc, err)
}

func (c *coordinator) attachConn(pc *peerConn, err error) (PeerConn, error) {
	if err != nil || pc == nil {
		return nil, err
	}
	c.trackConn(pc)
	return &trackedConn{peerConn: pc, untrack: func() { c.untrackConn(pc.remotePeerID) }}, nil
}

func (c *coordinator) trackConn(pc *peerConn) {
	c.connMu.Lock()
	defer c.connMu.Unlock()
	c.active[pc.remotePeerID] = ConnectionStats{
		RemotePeerID: pc.remotePeerID,
		Pair:         pc.pair,
	}
}

func (c *coordinator) untrackConn(remotePeerID string) {
	c.connMu.Lock()
	defer c.connMu.Unlock()
	delete(c.active, remotePeerID)
}

func (c *coordinator) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return nil
	}
	c.closed = true
	close(c.stopCh)
	return nil
}

func (c *coordinator) acceptLoop(ctx context.Context) {
	for {
		select {
		case <-c.stopCh:
			return
		case <-ctx.Done():
			return
		default:
		}

		iceCtx, cancel := context.WithTimeout(ctx, c.cfg.ICETimeout)
		iceRes, err := ice.Connect(iceCtx, toICEConfig(c.cfg), c.broker, c.selfID, "", ice.RoleListener)
		cancel()
		if err != nil {
			select {
			case <-c.stopCh:
				return
			case <-ctx.Done():
				return
			default:
				continue
			}
		}
		quicCtx, qcancel := context.WithTimeout(ctx, c.cfg.ICETimeout)
		pc, err := c.quicAccept(quicCtx, iceRes)
		qcancel()
		if err != nil {
			_ = iceRes.ICE.Close()
			_ = iceRes.Agent.Close()
			select {
			case <-c.stopCh:
				return
			case <-ctx.Done():
				return
			default:
				continue
			}
		}

		select {
		case c.inbound <- pc:
		case <-c.stopCh:
			_ = pc.Close()
			return
		}
	}
}

type trackedConn struct {
	peerConn *peerConn
	untrack  func()
}

func (t *trackedConn) Open(ctx context.Context) (tunnel.Stream, error) { return t.peerConn.Open(ctx) }
func (t *trackedConn) OpenWithPriority(ctx context.Context, pri int) (tunnel.Stream, error) {
	return t.peerConn.OpenWithPriority(ctx, pri)
}
func (t *trackedConn) Accept() (tunnel.Stream, error) { return t.peerConn.Accept() }
func (t *trackedConn) Close() error {
	err := t.peerConn.Close()
	if t.untrack != nil {
		t.untrack()
	}
	return err
}
func (t *trackedConn) IsClosed() bool                    { return t.peerConn.IsClosed() }
func (t *trackedConn) NumStreams() int                   { return t.peerConn.NumStreams() }
func (t *trackedConn) LocalAddr() net.Addr               { return t.peerConn.LocalAddr() }
func (t *trackedConn) RemoteAddr() net.Addr              { return t.peerConn.RemoteAddr() }
func (t *trackedConn) RemotePeerID() string              { return t.peerConn.RemotePeerID() }
func (t *trackedConn) LocalPeerID() string               { return t.peerConn.LocalPeerID() }
func (t *trackedConn) SelectedPair() CandidatePairInfo { return t.peerConn.SelectedPair() }

func (c *coordinator) quicDial(ctx context.Context, iceRes *ice.Connection) (*peerConn, error) {
	pc, remote, err := ice.NewPacketConn(iceRes.ICE)
	if err != nil {
		_ = iceRes.ICE.Close()
		_ = iceRes.Agent.Close()
		return nil, err
	}

	sess, err := quicconn.DialSession(ctx, pc, remote, c.transOpt)
	if err != nil {
		_ = pc.Close()
		_ = iceRes.Agent.Close()
		return nil, err
	}

	pair, _ := ice.SelectedPairInfo(iceRes.Agent)
	return &peerConn{
		session:      sess,
		localPeerID:  c.selfID,
		remotePeerID: iceRes.RemotePeerID,
		pair:         toCandidatePair(pair),
		iceAgent:     iceRes.Agent,
		iceConn:      iceRes.ICE,
		packetConn:   pc,
	}, nil
}

func (c *coordinator) quicAccept(ctx context.Context, iceRes *ice.Connection) (*peerConn, error) {
	pkt, _, err := ice.NewPacketConn(iceRes.ICE)
	if err != nil {
		return nil, err
	}

	sess, err := quicconn.AcceptOne(ctx, pkt, c.transOpt)
	if err != nil {
		_ = pkt.Close()
		return nil, err
	}

	pair, _ := ice.SelectedPairInfo(iceRes.Agent)
	return &peerConn{
		session:      sess,
		localPeerID:  c.selfID,
		remotePeerID: iceRes.RemotePeerID,
		pair:         toCandidatePair(pair),
		iceAgent:     iceRes.Agent,
		iceConn:      iceRes.ICE,
		packetConn:   pkt,
	}, nil
}

// Stats 返回当前节点摘要（供 debug 使用）。
func (c *coordinator) Stats() Stats {
	c.mu.Lock()
	listening := c.listener != nil && !c.closed
	selfID := c.selfID
	c.mu.Unlock()

	c.connMu.Lock()
	conns := make([]ConnectionStats, 0, len(c.active))
	for _, st := range c.active {
		conns = append(conns, st)
	}
	c.connMu.Unlock()

	return Stats{
		SelfID:      selfID,
		Listening:   listening,
		Connections: conns,
	}
}

type tunnelListener struct {
	coord *coordinator
	ctx   context.Context
}

func (l *tunnelListener) Accept(ctx context.Context) (PeerConn, error) {
	if ctx == nil {
		ctx = l.ctx
	}
	select {
	case pc := <-l.coord.inbound:
		return l.coord.attachConn(pc, nil)
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-l.coord.stopCh:
		return nil, ErrSignalingClosed
	}
}

func (l *tunnelListener) Close() error {
	return l.coord.Close()
}

type peerConn struct {
	session      tunnel.Session
	localPeerID  string
	remotePeerID string
	pair         CandidatePairInfo
	iceAgent     *pionice.Agent
	iceConn      *pionice.Conn
	packetConn   net.PacketConn
}

func (p *peerConn) Open(ctx context.Context) (tunnel.Stream, error) { return p.session.Open(ctx) }
func (p *peerConn) OpenWithPriority(ctx context.Context, pri int) (tunnel.Stream, error) {
	return p.session.OpenWithPriority(ctx, pri)
}
func (p *peerConn) Accept() (tunnel.Stream, error) { return p.session.Accept() }
func (p *peerConn) Close() error {
	if p.session != nil {
		_ = p.session.Close()
	}
	if p.packetConn != nil {
		_ = p.packetConn.Close()
	}
	if p.iceAgent != nil {
		_ = p.iceAgent.Close()
	}
	return nil
}
func (p *peerConn) IsClosed() bool                    { return p.session != nil && p.session.IsClosed() }
func (p *peerConn) NumStreams() int                   { return p.session.NumStreams() }
func (p *peerConn) LocalAddr() net.Addr               { return p.session.LocalAddr() }
func (p *peerConn) RemoteAddr() net.Addr              { return p.session.RemoteAddr() }
func (p *peerConn) RemotePeerID() string              { return p.remotePeerID }
func (p *peerConn) LocalPeerID() string               { return p.localPeerID }
func (p *peerConn) SelectedPair() CandidatePairInfo { return p.pair }

var (
	_ Coordinator = (*coordinator)(nil)
	_ Listener    = (*tunnelListener)(nil)
	_ PeerConn    = (*peerConn)(nil)
	_ PeerConn    = (*trackedConn)(nil)
	_ io.Closer   = (*peerConn)(nil)
	_ io.Closer   = (*trackedConn)(nil)
)