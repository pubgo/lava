package p2p

import (
	"context"
	"io"
	"net"
	"sync"
	"time"

	pionice "github.com/pion/ice/v4"

	"github.com/pubgo/lava/v2/core/p2p/ice"
	"github.com/pubgo/lava/v2/core/p2p/quicconn"
	"github.com/pubgo/lava/v2/core/p2p/signaling"
	"github.com/pubgo/lava/v2/core/tunnel"
)

type coordinator struct {
	cfg      Config
	broker   signaling.Broker
	sigMux   *signaling.Multiplex
	selfID   string
	transOpt *tunnel.TransportOptions
	metrics  *MetricsRecorder

	mu       sync.Mutex
	closed   bool
	inbound  chan *peerConn
	stopCh   chan struct{}
	listener *tunnelListener
	connMu   sync.Mutex
	active   map[string]*trackedPeerConn
}

type trackedPeerConn struct {
	stats ConnectionStats
	pc    *peerConn
}

// NewCoordinator 创建 P2P 协调器。
func NewCoordinator(cfg Config, broker signaling.Broker, selfID string) Coordinator {
	return NewCoordinatorWithMetrics(cfg, broker, selfID, nil)
}

// NewCoordinatorWithMetrics 创建 P2P 协调器并绑定可选 metrics。
func NewCoordinatorWithMetrics(cfg Config, broker signaling.Broker, selfID string, met *MetricsRecorder) Coordinator {
	if cfg.ICETimeout <= 0 {
		cfg.ICETimeout = DefaultConfig().ICETimeout
	}
	return &coordinator{
		cfg:      cfg,
		broker:   broker,
		sigMux:   signaling.NewMultiplex(broker, selfID),
		selfID:   selfID,
		transOpt: cfg.TransportOptions(),
		metrics:  met,
		inbound:  make(chan *peerConn),
		stopCh:   make(chan struct{}),
		active:   make(map[string]*trackedPeerConn),
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
	pc, err := c.dialOnce(ctx, peerID)
	if err != nil {
		if c.metrics != nil {
			c.metrics.ObserveDialFailure()
		}
		return nil, err
	}
	return c.attachConn(pc, nil)
}

func (c *coordinator) Reconnect(ctx context.Context, peerID string) (PeerConn, error) {
	if peerID == "" {
		return nil, ErrPeerNotFound
	}
	c.closeActivePeer(peerID)
	return c.dialWithRetry(ctx, peerID)
}

func (c *coordinator) dialOnce(ctx context.Context, peerID string) (*peerConn, error) {
	iceCtx, cancel := context.WithTimeout(ctx, c.cfg.ICETimeout)
	defer cancel()

	iceRes, err := c.iceConnect(iceCtx, peerID, ice.RoleDialer)
	if err != nil {
		return nil, err
	}
	return c.quicDial(iceCtx, iceRes)
}

func (c *coordinator) dialWithRetry(ctx context.Context, peerID string) (PeerConn, error) {
	max := c.cfg.Reconnect.MaxAttempts
	if max <= 0 {
		max = 3
	}
	backoff := c.cfg.Reconnect.Backoff
	if backoff <= 0 {
		backoff = time.Second
	}

	var lastErr error
	for attempt := 1; attempt <= max; attempt++ {
		if attempt > 1 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoff):
			}
		}
		pc, err := c.dialOnce(ctx, peerID)
		if err == nil {
			return c.attachConn(pc, nil)
		}
		lastErr = err
		if c.metrics != nil {
			c.metrics.ObserveDialFailure()
		}
	}
	if lastErr == nil {
		lastErr = ErrICEFailed
	}
	return nil, lastErr
}

func (c *coordinator) closeActivePeer(peerID string) {
	c.connMu.Lock()
	tracked, ok := c.active[peerID]
	if ok {
		delete(c.active, peerID)
	}
	c.connMu.Unlock()
	if ok && tracked.pc != nil {
		_ = tracked.pc.Close()
	}
	if c.metrics != nil {
		c.updateMetricsGauges()
	}
}

func (c *coordinator) attachConn(pc *peerConn, err error) (PeerConn, error) {
	if err != nil || pc == nil {
		return nil, err
	}
	c.trackConn(pc)
	return &trackedConn{peerConn: pc, untrack: func() { c.untrackConn(pc.remotePeerID) }}, nil
}

func (c *coordinator) trackConn(pc *peerConn) {
	st := ConnectionStats{
		RemotePeerID:      pc.remotePeerID,
		Pair:              pc.pair,
		ConnectedAt:       time.Now().UTC(),
		ConnectDurationMs: pc.connectDuration.Milliseconds(),
	}
	c.enrichLiveStats(&st, pc.iceAgent)

	c.connMu.Lock()
	c.active[pc.remotePeerID] = &trackedPeerConn{stats: st, pc: pc}
	c.connMu.Unlock()

	if c.metrics != nil {
		c.metrics.ObserveConnect(st.Pair, pc.connectDuration)
		c.updateMetricsGauges()
	}
}

func (c *coordinator) untrackConn(remotePeerID string) {
	c.connMu.Lock()
	delete(c.active, remotePeerID)
	c.connMu.Unlock()
	if c.metrics != nil {
		c.updateMetricsGauges()
	}
}

func (c *coordinator) enrichLiveStats(st *ConnectionStats, agent *pionice.Agent) {
	if live, ok := ice.LivePairStatsFromAgent(agent); ok {
		st.Pair = toCandidatePair(live.Pair)
		st.RTTMs = live.RTTMs
		st.PairState = live.PairState
		st.Nominated = live.Nominated
	}
}

func (c *coordinator) updateMetricsGauges() {
	if c.metrics == nil {
		return
	}
	c.connMu.Lock()
	active := len(c.active)
	relay := 0
	for _, t := range c.active {
		if isRelayPair(t.stats.Pair) {
			relay++
		}
	}
	c.connMu.Unlock()
	c.metrics.SetActiveConnections(active, relay)
}

func isRelayPair(pair CandidatePairInfo) bool {
	return pair.LocalType == "relay" || pair.RemoteType == "relay"
}

func (c *coordinator) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return nil
	}
	c.closed = true
	close(c.stopCh)
	if c.sigMux != nil {
		_ = c.sigMux.Close()
	}
	return nil
}

func (c *coordinator) iceConnect(ctx context.Context, peerID string, role ice.Role) (*ice.Connection, error) {
	sess := c.sigMux.Session()
	defer sess.Close()
	iceCfg, err := toICEConfig(c.cfg, c.selfID)
	if err != nil {
		return nil, err
	}
	return ice.Connect(ctx, iceCfg, sess, c.selfID, peerID, role)
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
		iceRes, err := c.iceConnect(iceCtx, "", ice.RoleListener)
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
		session:         sess,
		localPeerID:     c.selfID,
		remotePeerID:    iceRes.RemotePeerID,
		pair:            toCandidatePair(pair),
		connectDuration: iceRes.ConnectDuration,
		iceAgent:        iceRes.Agent,
		iceConn:         iceRes.ICE,
		packetConn:      pc,
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
		session:         sess,
		localPeerID:     c.selfID,
		remotePeerID:    iceRes.RemotePeerID,
		pair:            toCandidatePair(pair),
		connectDuration: iceRes.ConnectDuration,
		iceAgent:        iceRes.Agent,
		iceConn:         iceRes.ICE,
		packetConn:      pkt,
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
	relay := 0
	for _, t := range c.active {
		st := t.stats
		c.enrichLiveStats(&st, t.pc.iceAgent)
		conns = append(conns, st)
		if isRelayPair(st.Pair) {
			relay++
		}
	}
	active := len(conns)
	c.connMu.Unlock()

	return Stats{
		SelfID:            selfID,
		Listening:         listening,
		ActiveConnections: active,
		RelayConnections:  relay,
		Connections:       conns,
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
	session         tunnel.Session
	localPeerID     string
	remotePeerID    string
	pair            CandidatePairInfo
	connectDuration time.Duration
	iceAgent        *pionice.Agent
	iceConn         *pionice.Conn
	packetConn      net.PacketConn
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