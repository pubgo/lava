package p2p_test

import (
	"context"
	"io"
	"testing"
	"time"

	"github.com/pubgo/lava/v2/core/p2p"
	"github.com/pubgo/lava/v2/core/p2p/signaling"
)

// TestDevSTUNCoordinatorQUIC 经 dev coturn STUN/TURN 完成 ICE+QUIC 全链路（需外网）。
func TestDevSTUNCoordinatorQUIC(t *testing.T) {
	if testing.Short() {
		t.Skip("integration test")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	cfg := p2p.DefaultConfig()
	cfg.Insecure = true
	cfg.ICETimeout = 45 * time.Second

	brokerA, brokerB := signaling.Pair()
	coordA := p2p.NewCoordinator(cfg, brokerA, "dev-a")
	coordB := p2p.NewCoordinator(cfg, brokerB, "dev-b")
	t.Cleanup(func() {
		_ = coordA.Close()
		_ = coordB.Close()
	})

	ln, err := coordB.Listen(ctx, "dev-b")
	if err != nil {
		t.Fatal(err)
	}

	type dialResult struct {
		pc  p2p.PeerConn
		err error
	}
	dialCh := make(chan dialResult, 1)
	go func() {
		pc, err := coordA.Dial(ctx, "dev-b")
		dialCh <- dialResult{pc: pc, err: err}
	}()

	peerB, err := ln.Accept(ctx)
	if err != nil {
		t.Fatal(err)
	}
	dr := <-dialCh
	if dr.err != nil {
		t.Fatal(dr.err)
	}
	peerA := dr.pc
	defer peerA.Close()
	defer peerB.Close()

	pair := peerA.SelectedPair()
	t.Logf("ICE pair: local=%s/%s remote=%s/%s", pair.LocalType, pair.LocalAddr, pair.RemoteType, pair.RemoteAddr)

	readCh := make(chan string, 1)
	go func() {
		st, err := peerB.Accept()
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

	stA, err := peerA.Open(ctx)
	if err != nil {
		t.Fatal(err)
	}
	want := "dev-stun-quic"
	if _, err := stA.Write([]byte(want)); err != nil {
		t.Fatal(err)
	}
	if got := <-readCh; got != want {
		t.Fatalf("got %q", got)
	}
}
