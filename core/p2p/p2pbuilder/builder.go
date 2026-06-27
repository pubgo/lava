// Package p2pbuilder 装配 P2P Coordinator：tunnel agent 信令、lifecycle 与 debug 端点。
package p2pbuilder

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/pubgo/funk/v2/merge"
	lo "github.com/samber/lo"

	"github.com/pubgo/lava/v2/core/lifecycle"
	"github.com/pubgo/lava/v2/core/p2p"
	"github.com/pubgo/lava/v2/core/p2p/p2pdebug"
	"github.com/pubgo/lava/v2/core/p2p/signaling/tunnelsig"
	p2ptransport "github.com/pubgo/lava/v2/core/p2p/transport"
	"github.com/pubgo/lava/v2/core/tunnel/tunnelagent"
)

const Name = "p2p"

// Params DI 装配参数。
type Params struct {
	LC     lifecycle.Lifecycle
	Config *p2p.Config
	Agent  *tunnelagent.Agent
	PeerID string
	// ListenOnStart 为 true 时在 AfterStart 自动 Listen（agent 入站 P2P）。
	ListenOnStart bool
}

type node struct {
	mu     sync.RWMutex
	params Params
	cfg    p2p.Config
	coord  p2p.Coordinator
	broker *tunnelsig.Broker
}

// New 创建 P2P 节点：AfterStart 经 tunnel agent 注册信令并 Listen；BeforeStop 释放资源。
func New(p Params) (p2p.Coordinator, error) {
	if p.Agent == nil {
		return nil, fmt.Errorf("p2p: tunnel agent required")
	}
	if p.PeerID == "" {
		return nil, p2p.ErrPeerNotFound
	}
	if p.LC == nil {
		return nil, fmt.Errorf("p2p: lifecycle required")
	}

	cfg := lo.FromPtr(merge.Struct(lo.ToPtr(p2p.DefaultConfig()), p.Config).Unwrap())
	n := &node{params: p, cfg: cfg}

	p.LC.AfterStart(n.start)
	p.LC.BeforeStop(n.stop)
	return n, nil
}

func (n *node) start(ctx context.Context) error {
	if err := waitAgentSession(ctx, n.params.Agent); err != nil {
		return err
	}
	sess := n.params.Agent.Session()
	if sess == nil {
		return fmt.Errorf("p2p: tunnel agent session not ready")
	}

	broker := tunnelsig.New(sess, n.params.PeerID, n.cfg.AuthToken)
	tunnelsig.AttachAgentHandler(n.params.Agent.Config(), broker)
	if err := broker.Start(ctx); err != nil {
		return err
	}

	coord := p2p.NewCoordinator(n.cfg, broker, n.params.PeerID)

	n.mu.Lock()
	n.coord = coord
	n.broker = broker
	n.mu.Unlock()

	p2ptransport.SetCoordinator(coord)
	p2pdebug.Register(coord)

	if n.params.ListenOnStart {
		if _, err := coord.Listen(ctx, n.params.PeerID); err != nil {
			return err
		}
	}
	return nil
}

func (n *node) stop(ctx context.Context) error {
	n.mu.Lock()
	coord, broker := n.coord, n.broker
	n.coord, n.broker = nil, nil
	n.mu.Unlock()

	if coord != nil {
		p2ptransport.ClearCoordinator(coord)
	} else {
		p2ptransport.ClearCoordinator(nil)
	}
	p2pdebug.Register(nil)
	if coord != nil {
		_ = coord.Close()
	}
	if broker != nil {
		return broker.Close()
	}
	return nil
}

func (n *node) coordOrErr() (p2p.Coordinator, error) {
	n.mu.RLock()
	defer n.mu.RUnlock()
	if n.coord == nil {
		return nil, p2p.ErrSignalingClosed
	}
	return n.coord, nil
}

func (n *node) Listen(ctx context.Context, selfID string) (p2p.Listener, error) {
	c, err := n.coordOrErr()
	if err != nil {
		return nil, err
	}
	return c.Listen(ctx, selfID)
}

func (n *node) Dial(ctx context.Context, peerID string) (p2p.PeerConn, error) {
	c, err := n.coordOrErr()
	if err != nil {
		return nil, err
	}
	return c.Dial(ctx, peerID)
}

func (n *node) Close() error {
	return n.stop(context.Background())
}

func (n *node) Stats() p2p.Stats {
	n.mu.RLock()
	c := n.coord
	peerID := n.params.PeerID
	n.mu.RUnlock()
	if c == nil {
		return p2p.Stats{SelfID: peerID}
	}
	return c.Stats()
}

var _ p2p.Coordinator = (*node)(nil)

const agentSessionWait = 30 * time.Second

func waitAgentSession(ctx context.Context, agent *tunnelagent.Agent) error {
	if agent == nil {
		return fmt.Errorf("p2p: tunnel agent required")
	}
	if agent.Session() != nil {
		return nil
	}

	waitCtx := ctx
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		waitCtx, cancel = context.WithTimeout(ctx, agentSessionWait)
		defer cancel()
	}

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	for {
		if agent.Session() != nil {
			return nil
		}
		select {
		case <-waitCtx.Done():
			return fmt.Errorf("p2p: tunnel agent session not ready: %w", waitCtx.Err())
		case <-ticker.C:
		}
	}
}
