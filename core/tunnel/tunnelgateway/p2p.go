package tunnelgateway

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/pubgo/funk/v2/log"

	"github.com/pubgo/lava/v2/core/p2p/signaling"
	"github.com/pubgo/lava/v2/core/tunnel"
)

type peerListEntry struct {
	PeerID       string    `json:"peer_id"`
	AgentID      string    `json:"agent_id"`
	RegisteredAt time.Time `json:"registered_at"`
}

// handlePeerList 返回已注册 P2P peer 列表（gateway 侧 debug）。
func (g *tunnelGateway) handlePeerList(w http.ResponseWriter, r *http.Request) {
	g.mu.RLock()
	peers := make([]peerListEntry, 0, len(g.peers))
	for _, p := range g.peers {
		peers = append(peers, peerListEntry{
			PeerID:       p.peerID,
			AgentID:      p.agentID,
			RegisteredAt: p.registeredAt,
		})
	}
	g.mu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]any{
		"peers": peers,
		"count": len(peers),
	}); err != nil {
		log.Warn().Err(err).Msg("Gateway: failed to encode peer list")
	}
}

func (g *tunnelGateway) handleP2PRegister(agentID string, session tunnel.Session, msg *tunnel.Message) {
	if !g.p2pRegisterLimiter.Allow(agentID) {
		log.Warn().Str("agent", agentID).Msg("P2P register: rate limit exceeded")
		return
	}
	payload, err := tunnel.DecodeP2PRegisterPayload(msg.Payload)
	if err != nil {
		log.Warn().Err(err).Str("agent", agentID).Msg("P2P register: invalid payload")
		return
	}
	if payload.PeerID == "" {
		log.Warn().Str("agent", agentID).Msg("P2P register: empty peer_id")
		return
	}
	if g.authProvider != nil {
		if payload.AuthToken == "" {
			log.Warn().Str("peer", payload.PeerID).Msg("P2P register: missing auth_token")
			return
		}
		if _, err := g.authProvider.ValidateToken(payload.AuthToken); err != nil {
			log.Warn().Err(err).Str("peer", payload.PeerID).Msg("P2P register: auth failed")
			return
		}
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	if oldAgent, ok := g.peerByAgent[agentID]; ok && oldAgent != payload.PeerID {
		delete(g.peers, oldAgent)
	}
	if existing, ok := g.peers[payload.PeerID]; ok && existing.agentID != agentID {
		log.Warn().Str("peer", payload.PeerID).Str("agent", agentID).Msg("P2P register: peer_id already taken")
		return
	}

	g.peers[payload.PeerID] = &registeredPeer{
		peerID:       payload.PeerID,
		session:      session,
		agentID:      agentID,
		registeredAt: time.Now().UTC(),
	}
	g.peerByAgent[agentID] = payload.PeerID
	log.Info().Str("peer", payload.PeerID).Str("agent", agentID).Msg("P2P peer registered")
}

func (g *tunnelGateway) handleP2PSignal(agentID string, session tunnel.Session, msg *tunnel.Message) {
	var sig signaling.Message
	if err := json.Unmarshal(msg.Payload, &sig); err != nil {
		log.Warn().Err(err).Str("agent", agentID).Msg("P2P signal: invalid payload")
		return
	}
	if sig.To == "" {
		log.Warn().Str("agent", agentID).Msg("P2P signal: missing recipient")
		return
	}

	g.mu.RLock()
	senderPeer, senderOK := g.peerByAgent[agentID]
	if !senderOK {
		g.mu.RUnlock()
		log.Warn().Str("agent", agentID).Msg("P2P signal: sender not registered")
		return
	}
	if sig.From != "" && sig.From != senderPeer {
		g.mu.RUnlock()
		log.Warn().Str("agent", agentID).Str("from", sig.From).Str("registered", senderPeer).Msg("P2P signal: from mismatch")
		return
	}
	sig.From = senderPeer
	if !g.p2pSignalLimiter.Allow(senderPeer) {
		log.Warn().Str("from", senderPeer).Msg("P2P signal: rate limit exceeded")
		return
	}
	target, ok := g.peers[sig.To]
	g.mu.RUnlock()

	if !ok {
		log.Warn().Str("to", sig.To).Msg("P2P signal: peer not found")
		return
	}
	if g.authProvider != nil {
		if sig.AuthToken == "" {
			log.Warn().Str("from", sig.From).Msg("P2P signal: missing auth_token")
			return
		}
		if _, err := g.authProvider.ValidateToken(sig.AuthToken); err != nil {
			log.Warn().Err(err).Str("from", sig.From).Msg("P2P signal: auth failed")
			return
		}
	}

	payload, err := json.Marshal(sig)
	if err != nil {
		log.Warn().Err(err).Msg("P2P signal: marshal failed")
		return
	}
	g.deliverP2PSignal(target, payload)
}

func (g *tunnelGateway) deliverP2PSignal(target *registeredPeer, payload []byte) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	stream, err := target.session.Open(ctx)
	if err != nil {
		log.Warn().Err(err).Str("peer", target.peerID).Msg("P2P signal: open stream to peer failed")
		return
	}
	defer func() {
		if err := stream.Close(); err != nil {
			log.Warn().Err(err).Str("peer", target.peerID).Msg("P2P signal: close stream failed")
		}
	}()

	if err := tunnel.WriteMessage(stream, &tunnel.Message{
		Type:    tunnel.MessageTypeP2PSignal,
		Payload: payload,
	}); err != nil {
		log.Warn().Err(err).Str("peer", target.peerID).Msg("P2P signal: write failed")
	}
}

func (g *tunnelGateway) removePeerLocked(agentID string) {
	peerID, ok := g.peerByAgent[agentID]
	if !ok {
		return
	}
	delete(g.peerByAgent, agentID)
	delete(g.peers, peerID)
	log.Info().Str("peer", peerID).Str("agent", agentID).Msg("P2P peer removed (agent disconnected)")
}
