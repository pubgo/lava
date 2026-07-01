package quicconn_test

import (
	"context"
	"crypto/tls"
	"testing"
	"time"

	"github.com/quic-go/quic-go"

	"github.com/pubgo/lava/v2/core/p2p/quicconn"
	"github.com/pubgo/lava/v2/core/tunnel"
)

// TestServerTLSAndQUIC 验证 quicconn 生成的 TLS 配置可完成 QUIC 握手。
func TestServerTLSAndQUIC(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	opts := &tunnel.TransportOptions{Insecure: true, MaxStreams: 32}
	serverTLS, err := quicconn.ServerTLS(opts)
	if err != nil {
		t.Fatal(err)
	}

	ln, err := quic.ListenAddr("127.0.0.1:0", serverTLS, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = ln.Close() }()

	sessCh := make(chan error, 1)
	go func() {
		_, err := ln.Accept(ctx)
		sessCh <- err
	}()

	clientTLS := &tls.Config{
		InsecureSkipVerify: true,
		NextProtos:         []string{"lava-p2p"},
	}
	client, err := quic.DialAddr(ctx, ln.Addr().String(), clientTLS, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = client.CloseWithError(0, "done") }()

	if err := <-sessCh; err != nil {
		t.Fatal(err)
	}
}
