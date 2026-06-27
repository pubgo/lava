package tunnel

import (
	"encoding/json"
	"testing"
)

func TestDecodeP2PRegisterPayload(t *testing.T) {
	data, err := json.Marshal(P2PRegisterPayload{PeerID: "peer-1", AuthToken: "tok"})
	if err != nil {
		t.Fatal(err)
	}
	p, err := DecodeP2PRegisterPayload(data)
	if err != nil {
		t.Fatal(err)
	}
	if p.PeerID != "peer-1" || p.AuthToken != "tok" {
		t.Fatalf("got %+v", p)
	}
}

func TestMessageTypeP2PConstants(t *testing.T) {
	if MessageTypeP2PRegister <= MessageTypeDebugRequest {
		t.Fatal("P2P message types should follow debug request")
	}
	if MessageTypeP2PSignal != MessageTypeP2PRegister+1 {
		t.Fatal("unexpected P2PSignal constant")
	}
}
