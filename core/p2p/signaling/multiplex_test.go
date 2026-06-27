package signaling_test

import (
	"context"
	"testing"
	"time"

	"github.com/pubgo/lava/v2/core/p2p/signaling"
)

func TestMultiplexConcurrentRecv(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	hub := signaling.NewMemoryBroker()
	mux := signaling.NewMultiplex(hub, "a")
	t.Cleanup(func() { _ = mux.Close() })

	s1 := mux.Session()
	s2 := mux.Session()
	t.Cleanup(func() { _ = s1.Close() })
	t.Cleanup(func() { _ = s2.Close() })

	_ = hub.Send(ctx, signaling.Message{Type: signaling.TypeOffer, From: "b", To: "a", Ufrag: "u", Pwd: "p"})

	got1, err := s1.Recv(ctx, "a")
	if err != nil {
		t.Fatal(err)
	}
	got2, err := s2.Recv(ctx, "a")
	if err != nil {
		t.Fatal(err)
	}
	if got1.Ufrag != "u" || got2.Ufrag != "u" {
		t.Fatalf("s1=%+v s2=%+v", got1, got2)
	}
}
