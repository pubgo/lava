// Package p2pbuilder 装配 P2P Coordinator：tunnel agent 信令、lifecycle 与 debug 端点。
package p2pbuilder

import (
	"context"
	"fmt"

	"github.com/pubgo/lava/v2/core/lifecycle/lifecyclebuilder"
	"github.com/pubgo/lava/v2/core/metrics"
	"github.com/pubgo/lava/v2/core/p2p"
	"github.com/pubgo/lava/v2/core/tunnel"
	"github.com/pubgo/lava/v2/core/tunnel/tunnelagent"
	_ "github.com/pubgo/lava/v2/core/tunnel/yamux" // 注册 yamux 传输
)

// StandaloneOptions 一次性装配 P2P 节点所需的参数（自建 tunnel agent）。
type StandaloneOptions struct {
	// GatewayAddr tunnel gateway 地址（必填）。
	GatewayAddr string
	// PeerID 本节点 P2P 标识（必填）。
	PeerID string
	// AuthToken tunnel/P2P 鉴权 token。
	AuthToken string
	// ServiceName agent 注册名，默认 "p2p-<PeerID>"。
	ServiceName string
	// Config P2P 配置，nil 时用 DefaultConfig。
	Config *p2p.Config
	// ListenOnStart 为 true 时启动后自动 Listen（接受入站 P2P）。
	ListenOnStart bool
	// Metric 可选指标上报。
	Metric metrics.Metric
}

// Node 是一次性装配出的 P2P 节点句柄，内含 agent + coordinator + client。
type Node struct {
	Agent       *tunnelagent.Agent
	Coordinator p2p.Coordinator
	Client      *p2p.Client
	stop        func()
}

// Standalone 一行装配可用的 P2P 节点：建 tunnel agent、连接 gateway、
// 启动信令与 coordinator，并返回带池化/自动重连的 Client。
//
// 适合脚本、测试与独立进程；接入 DI/supervisor 时请用 New。
// 使用完毕调用 Node.Close 释放资源。
func Standalone(ctx context.Context, opts StandaloneOptions) (*Node, error) {
	if opts.GatewayAddr == "" {
		return nil, fmt.Errorf("p2p: gateway addr required")
	}
	if opts.PeerID == "" {
		return nil, p2p.ErrPeerNotFound
	}
	serviceName := opts.ServiceName
	if serviceName == "" {
		serviceName = "p2p-" + opts.PeerID
	}

	agent := tunnelagent.New(&tunnel.AgentConfig{
		GatewayAddr: opts.GatewayAddr,
		Transport:   tunnel.TransportYamux,
		ServiceName: serviceName,
		Metadata:    tunnel.ApplyAuthTokenMetadata(nil, opts.AuthToken),
	})
	if err := agent.Start(ctx); err != nil {
		return nil, err
	}

	cfg := p2p.DefaultConfig()
	if opts.Config != nil {
		cfg = *opts.Config
	}
	if cfg.AuthToken == "" {
		cfg.AuthToken = opts.AuthToken
	}

	lc := lifecyclebuilder.New(nil)
	coord, err := New(Params{
		LC:            lc.Setter,
		Agent:         agent,
		PeerID:        opts.PeerID,
		Config:        &cfg,
		Metric:        opts.Metric,
		ListenOnStart: opts.ListenOnStart,
	})
	if err != nil {
		_ = agent.Stop(context.Background())
		return nil, err
	}
	for _, hook := range lc.Getter.GetAfterStarts() {
		if err := hook.Exec(ctx); err != nil {
			_ = agent.Stop(context.Background())
			return nil, err
		}
	}

	stop := func() {
		for _, hook := range lc.Getter.GetBeforeStops() {
			_ = hook.Exec(ctx)
		}
		_ = agent.Stop(context.Background())
	}

	return &Node{
		Agent:       agent,
		Coordinator: coord,
		Client:      p2p.NewClient(coord),
		stop:        stop,
	}, nil
}

// Close 释放 Node 持有的全部资源。
func (n *Node) Close() error {
	if n.Client != nil {
		_ = n.Client.Close()
	}
	if n.stop != nil {
		n.stop()
	}
	return nil
}
