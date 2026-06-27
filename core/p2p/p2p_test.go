package p2p_test

import (
	"context"
	"io"
	"testing"
	"time"

	"github.com/pubgo/lava/v2/core/p2p"
	"github.com/pubgo/lava/v2/core/p2p/ice"
	"github.com/pubgo/lava/v2/core/p2p/quicconn"
	"github.com/pubgo/lava/v2/core/p2p/signaling"
	"github.com/pubgo/lava/v2/core/tunnel"
	_ "github.com/pubgo/lava/v2/core/p2p/transport"
)

func TestICEThenQUIC(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	brokerA, brokerB := signaling.Pair()
	cfg := ice.Config{ICETimeout: 15 * time.Second}
	opts := &tunnel.TransportOptions{Insecure: true, MaxStreams: 32}

	errCh := make(chan error, 2)
	var resA, resB *ice.Connection

	go func() {
		r, err := ice.Connect(ctx, cfg, brokerA, "a", "b", ice.RoleDialer)
		if err != nil {
			errCh <- err
			return
		}
		resA = r
		errCh <- nil
	}()
	go func() {
		r, err := ice.Connect(ctx, cfg, brokerB, "b", "a", ice.RoleListener)
		if err != nil {
			errCh <- err
			return
		}
		resB = r
		errCh <- nil
	}()
	for i := 0; i < 2; i++ {
		if err := <-errCh; err != nil {
			t.Fatal(err)
		}
	}

	readCh := make(chan string, 1)
	go func() {
		pkt, _, err := ice.NewPacketConn(resB.ICE)
		if err != nil {
			readCh <- err.Error()
			return
		}
		sess, err := quicconn.AcceptOne(ctx, pkt, opts)
		if err != nil {
			readCh <- err.Error()
			return
		}
		st, err := sess.Accept()
		if err != nil {
			readCh <- err.Error()
			return
		}
		buf := make([]byte, 16)
		n, _ := st.Read(buf)
		readCh <- string(buf[:n])
	}()

	pktA, remote, err := ice.NewPacketConn(resA.ICE)
	if err != nil {
		t.Fatal(err)
	}
	sessA, err := quicconn.DialSession(ctx, pktA, remote, opts)
	if err != nil {
		t.Fatal(err)
	}
	stA, err := sessA.Open(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := stA.Write([]byte("hello-quic")); err != nil {
		t.Fatal(err)
	}
	if got := <-readCh; got != "hello-quic" {
		t.Fatalf("got %q", got)
	}
}

func TestCoordinatorDialOnly(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	brokerA, brokerB := signaling.Pair()
	cfg := p2p.Config{ICETimeout: 10 * time.Second, Insecure: true}

	coordB := p2p.NewCoordinator(cfg, brokerB, "node-b")
	_, err := coordB.Listen(ctx, "node-b")
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(50 * time.Millisecond)

	coordA := p2p.NewCoordinator(cfg, brokerA, "node-a")
	pc, err := coordA.Dial(ctx, "node-b")
	if err != nil {
		t.Fatal(err)
	}
	_ = pc.Close()
	_ = coordA.Close()
	_ = coordB.Close()
}

func TestCoordinatorP2PQUIC(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	hub := signaling.NewMemoryBroker()
	cfg := p2p.Config{ICETimeout: 15 * time.Second, Insecure: true}

	coordA := p2p.NewCoordinator(cfg, hub, "node-a")
	coordB := p2p.NewCoordinator(cfg, hub, "node-b")
	t.Cleanup(func() {
		_ = coordA.Close()
		_ = coordB.Close()
	})

	ln, err := coordB.Listen(ctx, "node-b")
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(50 * time.Millisecond)

	type dialResult struct {
		pc  p2p.PeerConn
		err error
	}
	dialCh := make(chan dialResult, 1)
	go func() {
		pc, err := coordA.Dial(ctx, "node-b")
		dialCh <- dialResult{pc: pc, err: err}
	}()

	acceptCtx, acceptCancel := context.WithTimeout(ctx, 10*time.Second)
	defer acceptCancel()
	peerB, err := ln.Accept(acceptCtx)
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

	if peerA.RemotePeerID() != "node-b" || peerB.RemotePeerID() != "node-a" {
		t.Fatalf("peer ids: a->%s b->%s", peerA.RemotePeerID(), peerB.RemotePeerID())
	}

	readCh := make(chan string, 1)
	go func() {
		streamB, err := peerB.Accept()
		if err != nil {
			readCh <- err.Error()
			return
		}
		buf := make([]byte, 32)
		n, err := streamB.Read(buf)
		if err != nil && err != io.EOF {
			readCh <- err.Error()
			return
		}
		readCh <- string(buf[:n])
	}()

	streamA, err := peerA.Open(ctx)
	if err != nil {
		t.Fatal(err)
	}

	msg := []byte("p2p-quic-stream")
	if _, err := streamA.Write(msg); err != nil {
		t.Fatal(err)
	}
	if got := <-readCh; got != string(msg) {
		t.Fatalf("got %q", got)
	}
}
