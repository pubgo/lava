package ice_test

import (
	"context"
	"io"
	"testing"
	"time"

	"github.com/pubgo/lava/v2/core/p2p/ice"
	"github.com/pubgo/lava/v2/core/p2p/signaling"
)

// TestConnectHostOnly 本机 host 候选直连（不依赖 STUN/TURN）。
func TestConnectHostOnly(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	brokerA, brokerB := signaling.Pair()
	cfg := ice.Config{ICETimeout: 10 * time.Second}

	errCh := make(chan error, 2)
	var connA interface{ Write([]byte) (int, error); Close() error }

	go func() {
		c, err := ice.Connect(ctx, cfg, brokerA, "a", "b", ice.RoleDialer)
		if err != nil {
			errCh <- err
			return
		}
		connA = c.ICE
		errCh <- nil
	}()

	var connB interface {
		Read([]byte) (int, error)
		Close() error
	}
	go func() {
		c, err := ice.Connect(ctx, cfg, brokerB, "b", "a", ice.RoleListener)
		if err != nil {
			errCh <- err
			return
		}
		connB = c.ICE
		errCh <- nil
	}()

	for i := 0; i < 2; i++ {
		if err := <-errCh; err != nil {
			t.Fatalf("connect: %v", err)
		}
	}

	payload := []byte("p2p-ice-ping")
	if _, err := connA.Write(payload); err != nil {
		t.Fatalf("write: %v", err)
	}
	buf := make([]byte, 64)
	n, err := connB.Read(buf)
	if err != nil && err != io.EOF {
		t.Fatalf("read: %v", err)
	}
	if string(buf[:n]) != string(payload) {
		t.Fatalf("got %q want %q", buf[:n], payload)
	}
	_ = connA.Close()
	_ = connB.Close()
}

// TestConnectListenerAnyPeer 模拟 Coordinator Listen：Listener 不预设对端 ID。
func TestConnectListenerAnyPeer(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	brokerA, brokerB := signaling.Pair()
	cfg := ice.Config{ICETimeout: 10 * time.Second}

	errCh := make(chan error, 2)
	go func() {
		c, err := ice.Connect(ctx, cfg, brokerA, "a", "b", ice.RoleDialer)
		if err != nil {
			errCh <- err
			return
		}
		_ = c.ICE.Close()
		_ = c.Agent.Close()
		errCh <- nil
	}()
	go func() {
		c, err := ice.Connect(ctx, cfg, brokerB, "b", "", ice.RoleListener)
		if err != nil {
			errCh <- err
			return
		}
		if c.RemotePeerID != "a" {
			errCh <- ice.ErrFailed
			return
		}
		_ = c.ICE.Close()
		_ = c.Agent.Close()
		errCh <- nil
	}()
	for i := 0; i < 2; i++ {
		if err := <-errCh; err != nil {
			t.Fatalf("connect: %v", err)
		}
	}
}

// TestConnectDevSTUN 经 dev coturn STUN 收集 srflx 候选（需外网）。
func TestConnectDevSTUN(t *testing.T) {
	if testing.Short() {
		t.Skip("integration test")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()

	brokerA, brokerB := signaling.Pair()
	cfg := ice.Config{
		STUNURLs:   []string{"stun:118.178.168.253:3478"},
		TURNURL:    "turn:118.178.168.253:3478",
		TURNUser:   "lava",
		TURNPass:   "lava-p2p-dev",
		ICETimeout: 30 * time.Second,
	}

	errCh := make(chan error, 2)
	go func() {
		c, err := ice.Connect(ctx, cfg, brokerA, "a", "b", ice.RoleDialer)
		if err != nil {
			errCh <- err
			return
		}
		_ = c.ICE.Close()
		_ = c.Agent.Close()
		errCh <- nil
	}()
	go func() {
		c, err := ice.Connect(ctx, cfg, brokerB, "b", "a", ice.RoleListener)
		if err != nil {
			errCh <- err
			return
		}
		_ = c.ICE.Close()
		_ = c.Agent.Close()
		errCh <- nil
	}()
	for i := 0; i < 2; i++ {
		if err := <-errCh; err != nil {
			t.Fatalf("connect with dev stun: %v", err)
		}
	}
}
