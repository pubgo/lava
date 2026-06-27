package ice

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	pionice "github.com/pion/ice/v4"
	"github.com/pion/logging"
	"github.com/pion/stun/v3"

	"github.com/pubgo/lava/v2/core/p2p/signaling"
)

// Connection ICE 协商结果（保留 Agent 以便读取选路与关闭资源）。
type Connection struct {
	ICE             *pionice.Conn
	Agent           *pionice.Agent
	RemotePeerID    string
	ConnectDuration time.Duration
}

// Connect 通过信令 Broker 与对端完成 ICE 协商，返回已建立的 ICE 连接。
func Connect(
	ctx context.Context,
	cfg Config,
	broker signaling.Broker,
	selfID, peerID string,
	role Role,
) (*Connection, error) {
	start := time.Now()
	urls, err := parseURLs(cfg)
	if err != nil {
		return nil, err
	}

	agentCfg := &pionice.AgentConfig{
		Urls:         urls,
		NetworkTypes: []pionice.NetworkType{pionice.NetworkTypeUDP4},
	}
	if lvl := os.Getenv("P2P_ICE_DEBUG"); lvl != "" {
		lf := logging.NewDefaultLoggerFactory()
		lf.DefaultLogLevel = logging.LogLevelTrace
		agentCfg.LoggerFactory = lf
	}
	agent, err := pionice.NewAgent(agentCfg)
	if err != nil {
		return nil, fmt.Errorf("ice agent: %w", err)
	}

	var (
		remoteReady sync.WaitGroup
		remoteMu    sync.Mutex
		remoteUfrag string
		remotePwd   string
		remoteSet   bool
		offerFrom   string
		offerMu     sync.Mutex
		pendingCand []string
		pendingMu   sync.Mutex
	)

	remoteReady.Add(1)

	peerTarget := func() string {
		if peerID != "" {
			return peerID
		}
		offerMu.Lock()
		defer offerMu.Unlock()
		return offerFrom
	}

	sendCandidate := func(raw string) {
		to := peerTarget()
		if to == "" {
			pendingMu.Lock()
			pendingCand = append(pendingCand, raw)
			pendingMu.Unlock()
			return
		}
		_ = broker.Send(ctx, signaling.Message{
			Type:      signaling.TypeCandidate,
			From:      selfID,
			To:        to,
			Candidate: raw,
			AuthToken: cfg.AuthToken,
		})
	}

	flushPendingCandidates := func() {
		pendingMu.Lock()
		pending := pendingCand
		pendingCand = nil
		pendingMu.Unlock()
		for _, raw := range pending {
			sendCandidate(raw)
		}
	}

	if err := agent.OnCandidate(func(c pionice.Candidate) {
		if c == nil {
			return
		}
		sendCandidate(c.Marshal())
	}); err != nil {
		_ = agent.Close()
		return nil, err
	}

	recvCtx, recvCancel := context.WithCancel(ctx)
	defer func() {
		if err != nil {
			recvCancel()
		}
	}()

	go func() {
		for {
			msg, err := broker.Recv(recvCtx, selfID)
			if err != nil {
				return
			}
			switch msg.Type {
			case signaling.TypeOffer, signaling.TypeAnswer:
				if peerID != "" && msg.From != peerID {
					continue
				}
				if msg.Type == signaling.TypeOffer {
					offerMu.Lock()
					if offerFrom == "" {
						offerFrom = msg.From
					}
					offerMu.Unlock()
					flushPendingCandidates()
				}
				remoteMu.Lock()
				if !remoteSet {
					remoteUfrag, remotePwd = msg.Ufrag, msg.Pwd
					remoteSet = true
					remoteMu.Unlock()
					_ = agent.SetRemoteCredentials(msg.Ufrag, msg.Pwd)
					remoteReady.Done()
				} else {
					remoteMu.Unlock()
				}
			case signaling.TypeCandidate:
				if peerID != "" && msg.From != peerID {
					continue
				}
				if msg.Candidate == "" {
					continue
				}
				cand, err := pionice.UnmarshalCandidate(msg.Candidate)
				if err != nil {
					continue
				}
				_ = agent.AddRemoteCandidate(cand)
			}
		}
	}()

	if err := agent.GatherCandidates(); err != nil {
		_ = agent.Close()
		return nil, fmt.Errorf("gather candidates: %w", err)
	}

	localUfrag, localPwd, err := agent.GetLocalUserCredentials()
	if err != nil {
		_ = agent.Close()
		return nil, err
	}

	switch role {
	case RoleDialer:
		if err := broker.Send(ctx, signaling.Message{
			Type: signaling.TypeOffer, From: selfID, To: peerID,
			Ufrag: localUfrag, Pwd: localPwd, AuthToken: cfg.AuthToken,
		}); err != nil {
			_ = agent.Close()
			return nil, err
		}
	case RoleListener:
		// 等待 Offer（由后台 goroutine 处理并 signal remoteReady）
	}

	waitDone := make(chan struct{})
	go func() {
		remoteReady.Wait()
		close(waitDone)
	}()

	select {
	case <-waitDone:
	case <-ctx.Done():
		_ = agent.Close()
		return nil, ctx.Err()
	}

	if role == RoleListener {
		to := peerTarget()
		if to == "" {
			_ = agent.Close()
			return nil, ErrFailed
		}
		if err := broker.Send(ctx, signaling.Message{
			Type: signaling.TypeAnswer, From: selfID, To: to,
			Ufrag: localUfrag, Pwd: localPwd, AuthToken: cfg.AuthToken,
		}); err != nil {
			_ = agent.Close()
			return nil, err
		}
		// Dialer 侧会收到 Answer；Listener 需再等 Answer 路径上已设 remote
	}

	remoteMu.Lock()
	ru, rp := remoteUfrag, remotePwd
	remoteMu.Unlock()
	if ru == "" || rp == "" {
		_ = agent.Close()
		return nil, ErrFailed
	}

	var iceConn *pionice.Conn
	remotePeerID := peerID
	switch role {
	case RoleDialer:
		iceConn, err = agent.Dial(ctx, ru, rp)
	default:
		offerMu.Lock()
		if remotePeerID == "" {
			remotePeerID = offerFrom
		}
		offerMu.Unlock()
		iceConn, err = agent.Accept(ctx, ru, rp)
	}
	if err != nil {
		_ = agent.Close()
		return nil, fmt.Errorf("ice connect: %w", err)
	}
	recvCancel()
	return &Connection{
		ICE:             iceConn,
		Agent:           agent,
		RemotePeerID:    remotePeerID,
		ConnectDuration: time.Since(start),
	}, nil
}

// SelectedPairInfo 从 agent 读取选路摘要。
func SelectedPairInfo(agent *pionice.Agent) (CandidatePairInfo, bool) {
	pair, err := agent.GetSelectedCandidatePair()
	if err != nil || pair == nil {
		return CandidatePairInfo{}, false
	}
	return CandidatePairInfo{
		LocalType:  pair.Local.Type().String(),
		RemoteType: pair.Remote.Type().String(),
		LocalAddr:  pair.Local.Address(),
		RemoteAddr: pair.Remote.Address(),
	}, true
}

func parseURLs(cfg Config) ([]*stun.URI, error) {
	var urls []*stun.URI
	for _, raw := range cfg.STUNURLs {
		u, err := stun.ParseURI(raw)
		if err != nil {
			return nil, fmt.Errorf("parse stun url %q: %w", raw, err)
		}
		urls = append(urls, u)
	}
	if cfg.TURNURL != "" {
		u, err := stun.ParseURI(cfg.TURNURL)
		if err != nil {
			return nil, fmt.Errorf("parse turn url %q: %w", cfg.TURNURL, err)
		}
		u.Username = cfg.TURNUser
		u.Password = cfg.TURNPass
		urls = append(urls, u)
	}
	return urls, nil
}
