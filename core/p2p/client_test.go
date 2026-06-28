package p2p_test

import (
	"context"
	"io"
	"testing"
	"time"

	"github.com/pubgo/lava/v2/core/p2p"
	"github.com/pubgo/lava/v2/core/p2p/signaling"
)

func TestClientPoolReuse(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cfg := p2p.Config{ICETimeout: 15 * time.Second, Insecure: true}
	brokerA, brokerB := signaling.Pair()
	coordA := p2p.NewCoordinator(cfg, brokerA, "node-a")
	coordB := p2p.NewCoordinator(cfg, brokerB, "node-b")
	defer coordA.Close()
	defer coordB.Close()

	ln, err := coordB.Listen(ctx, "node-b")
	if err != nil {
		t.Fatal(err)
	}

	acceptCh := make(chan p2p.PeerConn, 1)
	go func() {
		pc, err := ln.Accept(ctx)
		if err != nil {
			return
		}
		acceptCh <- pc
	}()

	client := p2p.NewClient(coordA)
	defer client.Close()

	pc1, err := client.Conn(ctx, "node-b")
	if err != nil {
		t.Fatal(err)
	}
	peerB := <-acceptCh
	if peerB == nil {
		t.Fatal("accept")
	}
	defer peerB.Close()

	pc2, err := client.Conn(ctx, "node-b")
	if err != nil {
		t.Fatal(err)
	}
	if pc1 != pc2 {
		t.Fatal("expected pooled connection reuse")
	}
}

func TestClientOpenStreamAfterClose(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()

	cfg := p2p.Config{
		ICETimeout: 15 * time.Second,
		Insecure:   true,
		Reconnect: p2p.ReconnectConfig{
			MaxAttempts: 3,
			Backoff:     50 * time.Millisecond,
		},
	}
	brokerA, brokerB := signaling.Pair()
	coordA := p2p.NewCoordinator(cfg, brokerA, "node-a")
	coordB := p2p.NewCoordinator(cfg, brokerB, "node-b")
	defer coordA.Close()
	defer coordB.Close()

	ln, err := coordB.Listen(ctx, "node-b")
	if err != nil {
		t.Fatal(err)
	}

	acceptCh := make(chan p2p.PeerConn, 2)
	go func() {
		for i := 0; i < 2; i++ {
			pc, err := ln.Accept(ctx)
			if err != nil {
				return
			}
			acceptCh <- pc
		}
	}()

	client := p2p.NewClient(coordA)
	defer client.Close()

	pc, err := client.Conn(ctx, "node-b")
	if err != nil {
		t.Fatal(err)
	}
	peerB1 := <-acceptCh
	if peerB1 == nil {
		t.Fatal("accept1")
	}
	if err := pc.Close(); err != nil {
		t.Fatal(err)
	}
	if err := peerB1.Close(); err != nil {
		t.Fatal(err)
	}

	readCh := make(chan string, 1)
	go func() {
		peerB2 := <-acceptCh
		if peerB2 == nil {
			readCh <- "accept failed"
			return
		}
		defer peerB2.Close()
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

	conn, err := client.OpenStream(ctx, "node-b")
	if err != nil {
		t.Fatal(err)
	}
	msg := []byte("auto-reconnect")
	if _, err := conn.Write(msg); err != nil {
		t.Fatal(err)
	}
	if got := <-readCh; got != string(msg) {
		t.Fatalf("got %q", got)
	}
}

func TestClientNetDialer(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cfg := p2p.Config{ICETimeout: 15 * time.Second, Insecure: true}
	brokerA, brokerB := signaling.Pair()
	coordA := p2p.NewCoordinator(cfg, brokerA, "node-a")
	coordB := p2p.NewCoordinator(cfg, brokerB, "node-b")
	defer coordA.Close()
	defer coordB.Close()

	ln, err := coordB.Listen(ctx, "node-b")
	if err != nil {
		t.Fatal(err)
	}

	acceptCh := make(chan p2p.PeerConn, 1)
	go func() {
		pc, err := ln.Accept(ctx)
		if err != nil {
			return
		}
		acceptCh <- pc
	}()

	client := p2p.NewClient(coordA)
	defer client.Close()

	dialer := client.NetDialer()
	conn, err := dialer(ctx, "node-b")
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	peerB := <-acceptCh
	if peerB == nil {
		t.Fatal("accept")
	}
	defer peerB.Close()

	readCh := make(chan string, 1)
	go func() {
		st, err := peerB.Accept()
		if err != nil {
			readCh <- err.Error()
			return
		}
		buf := make([]byte, 16)
		n, _ := st.Read(buf)
		readCh <- string(buf[:n])
	}()

	if _, err := conn.Write([]byte("via-dialer")); err != nil {
		t.Fatal(err)
	}
	if got := <-readCh; got != "via-dialer" {
		t.Fatalf("got %q", got)
	}
}
