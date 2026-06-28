package p2pbuilder_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/pubgo/lava/v2/core/p2p"
	"github.com/pubgo/lava/v2/core/p2p/p2pbuilder"
	"github.com/pubgo/lava/v2/core/tunnel"
	"github.com/pubgo/lava/v2/core/tunnel/tunnelgateway"
)

func TestStandaloneLocalPair(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()

	gwAddr := "127.0.0.1:27100"
	gw := tunnelgateway.NewGateway(&tunnel.GatewayConfig{
		ListenAddr: gwAddr,
		Transport:  tunnel.TransportYamux,
	})
	if err := tunnel.ConfigureGatewayAuth(gw, "test-token"); err != nil {
		t.Fatal(err)
	}
	if err := gw.Start(ctx); err != nil {
		t.Fatal(err)
	}
	defer gw.Stop(context.Background())

	cfg := p2p.Config{ICETimeout: 15 * time.Second, Insecure: true}
	nodeB, err := p2pbuilder.Standalone(ctx, p2pbuilder.StandaloneOptions{
		GatewayAddr:   gwAddr,
		PeerID:        "node-b",
		AuthToken:     "test-token",
		ListenOnStart: true,
		Config:        &cfg,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer nodeB.Close()

	nodeA, err := p2pbuilder.Standalone(ctx, p2pbuilder.StandaloneOptions{
		GatewayAddr: gwAddr,
		PeerID:      "node-a",
		AuthToken:   "test-token",
		Config:      &cfg,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer nodeA.Close()

	time.Sleep(200 * time.Millisecond)

	acceptCh := make(chan error, 1)
	go func() {
		ln, err := nodeB.Coordinator.Listen(ctx, "node-b")
		if err != nil {
			acceptCh <- err
			return
		}
		pc, err := ln.Accept(ctx)
		if err != nil {
			acceptCh <- err
			return
		}
		defer pc.Close()
		st, err := pc.Accept()
		if err != nil {
			acceptCh <- err
			return
		}
		buf := make([]byte, 16)
		n, err := st.Read(buf)
		if err != nil {
			acceptCh <- err
			return
		}
		if string(buf[:n]) != "standalone" {
			acceptCh <- fmt.Errorf("got %q", string(buf[:n]))
			return
		}
		acceptCh <- nil
	}()

	conn, err := nodeA.Client.OpenStream(ctx, "node-b")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := conn.Write([]byte("standalone")); err != nil {
		t.Fatal(err)
	}
	if err := <-acceptCh; err != nil {
		t.Fatal(err)
	}
}
