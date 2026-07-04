package testhelper

import (
	"context"
	"net"
	"sync/atomic"
	"testing"
	"time"
)

func TestWaitTCP(t *testing.T) {
	t.Parallel()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = ln.Close() }()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := WaitTCP(ctx, ln.Addr().String()); err != nil {
		t.Fatal(err)
	}
}

func TestWaitServiceCount(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	var n atomic.Int32
	go func() {
		time.Sleep(50 * time.Millisecond)
		n.Store(2)
	}()

	if err := WaitServiceCount(ctx, 2, func() int { return int(n.Load()) }); err != nil {
		t.Fatal(err)
	}
}
