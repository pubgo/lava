package tunnelgateway

import (
	"encoding/json"
	"testing"

	"github.com/pubgo/lava/v2/core/tunnel"
	"github.com/pubgo/lava/v2/core/tunnel/ratelimit"
)

func TestHandleP2PRegister(t *testing.T) {
	g := NewGateway(&tunnel.GatewayConfig{}).(*tunnelGateway)
	session := &fakeSession{}

	t.Run("empty_peer_id", func(t *testing.T) {
		reg, _ := json.Marshal(tunnel.P2PRegisterPayload{PeerID: ""})
		g.handleP2PRegister("agent-1", session, &tunnel.Message{Payload: reg})
		if len(g.peers) != 0 {
			t.Fatal("empty peer_id should not register")
		}
	})

	t.Run("success", func(t *testing.T) {
		reg, _ := json.Marshal(tunnel.P2PRegisterPayload{PeerID: "peer-1"})
		g.handleP2PRegister("agent-1", session, &tunnel.Message{Payload: reg})
		g.mu.RLock()
		p, ok := g.peers["peer-1"]
		g.mu.RUnlock()
		if !ok || p.agentID != "agent-1" {
			t.Fatalf("peer not registered: ok=%v", ok)
		}
	})

	t.Run("peer_taken", func(t *testing.T) {
		reg, _ := json.Marshal(tunnel.P2PRegisterPayload{PeerID: "peer-1"})
		g.handleP2PRegister("agent-2", &fakeSession{}, &tunnel.Message{Payload: reg})
		g.mu.RLock()
		p := g.peers["peer-1"]
		g.mu.RUnlock()
		if p.agentID != "agent-1" {
			t.Fatal("peer_id taken should keep original owner")
		}
	})

	t.Run("agent_reconnect_replaces", func(t *testing.T) {
		reg, _ := json.Marshal(tunnel.P2PRegisterPayload{PeerID: "peer-2"})
		g.handleP2PRegister("agent-1", session, &tunnel.Message{Payload: reg})
		g.mu.RLock()
		_, oldGone := g.peers["peer-1"]
		_, newOK := g.peers["peer-2"]
		g.mu.RUnlock()
		if oldGone {
			t.Fatal("old peer for same agent should be removed")
		}
		if !newOK {
			t.Fatal("new peer missing")
		}
	})
}

func TestHandleP2PRegister_RateLimit(t *testing.T) {
	g := NewGateway(&tunnel.GatewayConfig{P2PRegisterRateLimit: 1}).(*tunnelGateway)
	g.p2pRegisterLimiter = ratelimit.New(1)
	session := &fakeSession{}

	reg1, _ := json.Marshal(tunnel.P2PRegisterPayload{PeerID: "peer-a"})
	g.handleP2PRegister("agent-rl", session, &tunnel.Message{Payload: reg1})
	reg2, _ := json.Marshal(tunnel.P2PRegisterPayload{PeerID: "peer-b"})
	g.handleP2PRegister("agent-rl", session, &tunnel.Message{Payload: reg2})

	g.mu.RLock()
	_, hasA := g.peers["peer-a"]
	_, hasB := g.peers["peer-b"]
	g.mu.RUnlock()
	if !hasA {
		t.Fatal("first register should succeed")
	}
	if hasB {
		t.Fatal("second register should be rate limited")
	}
}
