package quicconn_test

import (
	"context"
	"testing"
	"time"

	"github.com/pubgo/lava/v2/core/tunnel"
	_ "github.com/pubgo/lava/v2/core/tunnel/quic"
)

func TestTunnelQUICListenAddr(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	tr, err := tunnel.NewTransport(tunnel.TransportQUIC, &tunnel.TransportOptions{Insecure: true})
	if err != nil {
		t.Fatal(err)
	}
	ln, err := tr.Listen(ctx, "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = ln.Close() }()

	sessCh := make(chan error, 1)
	go func() {
		s, err := ln.Accept()
		sessCh <- err
		_ = s
	}()

	client, err := tunnel.DialTransport(ctx, tunnel.TransportQUIC, ln.Addr().String(), &tunnel.TransportOptions{Insecure: true})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = client.Close() }()

	if err := <-sessCh; err != nil {
		t.Fatal(err)
	}
}
