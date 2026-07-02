package p2p_test

import (
	"context"
	"io"
	"testing"
	"time"

	"github.com/pubgo/lava/v2/core/p2p"
	"github.com/pubgo/lava/v2/core/p2p/signaling"
)

func TestCoordinatorReconnect(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()

	brokerA, brokerB := signaling.Pair()
	cfg := p2p.Config{
		ICETimeout: 10 * time.Second,
		Insecure:   true,
		Reconnect: p2p.ReconnectConfig{
			MaxAttempts: 3,
			Backoff:     50 * time.Millisecond,
		},
	}

	coordA := p2p.NewCoordinator(cfg, brokerA, "node-a")
	coordB := p2p.NewCoordinator(cfg, brokerB, "node-b")
	defer func() { _ = coordA.Close() }()
	defer func() { _ = coordB.Close() }()

	ln, err := coordB.Listen(ctx, "node-b")
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(50 * time.Millisecond)

	acceptCh := make(chan p2p.PeerConn, 2)
	go func() {
		for i := 0; i < 2; i++ {
			acceptCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
			pc, err := ln.Accept(acceptCtx)
			cancel()
			if err != nil {
				return
			}
			acceptCh <- pc
		}
	}()

	pc1, err := coordA.Dial(ctx, "node-b")
	if err != nil {
		t.Fatal(err)
	}
	peerB1 := <-acceptCh
	if peerB1 == nil {
		t.Fatal("accept peerB1")
	}
	if err := pc1.Close(); err != nil {
		t.Fatal(err)
	}
	if err := peerB1.Close(); err != nil {
		t.Fatal(err)
	}

	pc2, err := coordA.Reconnect(ctx, "node-b")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = pc2.Close() }()
	peerB2 := <-acceptCh
	if peerB2 == nil {
		t.Fatal("accept peerB2")
	}
	defer func() { _ = peerB2.Close() }()

	readCh := make(chan string, 1)
	go func() {
		st, err := peerB2.Accept()
		if err != nil {
			readCh <- err.Error()
			return
		}
		buf := make([]byte, 32)
		n, err := st.Read(buf)
		if err != nil && err != io.EOF {
			readCh <- err.Error()
			return
		}
		readCh <- string(buf[:n])
	}()

	st, err := pc2.Open(ctx)
	if err != nil {
		t.Fatal(err)
	}
	msg := []byte("after-reconnect")
	if _, err := st.Write(msg); err != nil {
		t.Fatal(err)
	}
	if got := <-readCh; got != string(msg) {
		t.Fatalf("got %q", got)
	}

	stA := coordA.Stats()
	if stA.ActiveConnections != 1 {
		t.Fatalf("active=%d want 1", stA.ActiveConnections)
	}
}
