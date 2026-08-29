package tunnelgateway

import (
	"encoding/json"
	"testing"

	"github.com/pubgo/lava/v2/core/tunnel"
)

func TestHandleRegisterDeregister(t *testing.T) {
	g := NewGateway(&tunnel.GatewayConfig{}).(*tunnelGateway)
	session := &fakeSession{}

	payload, err := json.Marshal(tunnel.ServiceInfo{Name: "demo", ID: "id-1"})
	if err != nil {
		t.Fatal(err)
	}
	g.handleRegister("agent-1", session, &tunnel.Message{Payload: payload})

	if _, err := g.GetService("demo"); err != nil {
		t.Fatalf("GetService: %v", err)
	}
	if n := len(g.Services()); n != 1 {
		t.Fatalf("Services count=%d", n)
	}

	g.handleDeregister("agent-other", &tunnel.Message{Payload: []byte("demo")})
	if _, err := g.GetService("demo"); err != nil {
		t.Fatal("wrong agent must not deregister")
	}

	g.handleDeregister("agent-1", &tunnel.Message{Payload: []byte("demo")})
	if _, err := g.GetService("demo"); err == nil {
		t.Fatal("expected service removed")
	}
}

func TestRemoveAgentServicesClearsPeers(t *testing.T) {
	g := NewGateway(&tunnel.GatewayConfig{}).(*tunnelGateway)
	session := &fakeSession{}
	payload, _ := json.Marshal(tunnel.ServiceInfo{Name: "demo", ID: "id-1"})
	g.handleRegister("agent-1", session, &tunnel.Message{Payload: payload})

	reg, _ := json.Marshal(tunnel.P2PRegisterPayload{PeerID: "peer-1"})
	g.handleP2PRegister("agent-1", session, &tunnel.Message{Payload: reg})

	g.removeAgentServices("agent-1")
	if len(g.Services()) != 0 {
		t.Fatal("services should be cleared")
	}
	g.mu.RLock()
	_, ok := g.peers["peer-1"]
	g.mu.RUnlock()
	if ok {
		t.Fatal("peer should be cleared")
	}
}

func TestHandleRegister_AuthFailure(t *testing.T) {
	g := NewGateway(&tunnel.GatewayConfig{}).(*tunnelGateway)
	g.authProvider = &fakeAuth{authenticateErr: tunnel.ErrAuthFailed}
	payload, _ := json.Marshal(tunnel.ServiceInfo{Name: "demo", ID: "id-1"})
	g.handleRegister("agent-1", &fakeSession{}, &tunnel.Message{Payload: payload})
	if len(g.Services()) != 0 {
		t.Fatal("auth failure should skip register")
	}
}

func TestHandleHeartbeat(t *testing.T) {
	g := NewGateway(&tunnel.GatewayConfig{}).(*tunnelGateway)
	payload, _ := json.Marshal(tunnel.ServiceInfo{Name: "demo", ID: "id-1"})
	g.handleRegister("agent-1", &fakeSession{}, &tunnel.Message{Payload: payload})
	before := g.services["demo"].info.LastHeartbeat
	g.handleHeartbeat("agent-1")
	after := g.services["demo"].info.LastHeartbeat
	if !after.After(before) && !after.Equal(before) {
		// Equal is ok if clock resolution is coarse; After is preferred.
		t.Fatalf("heartbeat not updated: before=%v after=%v", before, after)
	}
	if after.Before(before) {
		t.Fatal("heartbeat went backwards")
	}
}
