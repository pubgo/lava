package p2p_test

import (
	"context"
	"testing"
	"time"

	"github.com/pubgo/lava/v2/core/p2p"
	"github.com/pubgo/lava/v2/core/p2p/signaling"
)

func TestCoordinatorStatsFields(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	cfg := p2p.Config{ICETimeout: 10 * time.Second, Insecure: true}
	brokerA, brokerB := signaling.Pair()
	coordA := p2p.NewCoordinator(cfg, brokerA, "node-a")
	coordB := p2p.NewCoordinator(cfg, brokerB, "node-b")
	defer func() { _ = coordA.Close() }()
	defer func() { _ = coordB.Close() }()

	ln, err := coordB.Listen(ctx, "node-b")
	if err != nil {
		t.Fatal(err)
	}

	dialCh := make(chan p2p.PeerConn, 1)
	go func() {
		pc, err := coordA.Dial(ctx, "node-b")
		if err != nil {
			dialCh <- nil
			return
		}
		dialCh <- pc
	}()

	pcB, err := ln.Accept(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = pcB.Close() }()

	pcA := <-dialCh
	if pcA == nil {
		t.Fatal("dial failed")
	}
	defer func() { _ = pcA.Close() }()

	st := coordA.Stats()
	if st.ActiveConnections != 1 {
		t.Fatalf("active=%d want 1", st.ActiveConnections)
	}
	if len(st.Connections) != 1 {
		t.Fatalf("connections=%d want 1", len(st.Connections))
	}
	conn := st.Connections[0]
	if conn.ConnectDurationMs <= 0 {
		t.Fatalf("connect_duration_ms=%d want >0", conn.ConnectDurationMs)
	}
	if conn.Pair.LocalType == "" || conn.Pair.RemoteType == "" {
		t.Fatalf("pair types empty: %+v", conn.Pair)
	}
}
