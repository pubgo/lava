package tunnelgateway

import (
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/pubgo/lava/v2/core/tunnel"
	"github.com/pubgo/lava/v2/core/tunnel/ratelimit"
)

func TestHandleGRPCConnection_InvalidRoute(t *testing.T) {
	g := NewGateway(&tunnel.GatewayConfig{}).(*tunnelGateway)
	client, server := net.Pipe()
	t.Cleanup(func() { _ = client.Close() })

	done := make(chan struct{})
	go func() {
		defer close(done)
		g.handleGRPCConnection(server)
	}()

	if _, err := fmt.Fprintf(client, "NOT-A-ROUTE\n"); err != nil {
		t.Fatal(err)
	}
	_ = client.Close()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("handleGRPCConnection did not return")
	}
}

func TestHandleGRPCConnection_MissingService(t *testing.T) {
	g := NewGateway(&tunnel.GatewayConfig{}).(*tunnelGateway)
	client, server := net.Pipe()
	t.Cleanup(func() { _ = client.Close() })

	done := make(chan struct{})
	go func() {
		defer close(done)
		g.handleGRPCConnection(server)
	}()

	if _, err := fmt.Fprintf(client, "%smissing\n", tunnel.GRPCRoutePrefix); err != nil {
		t.Fatal(err)
	}
	// Forward fails with service not found; handler returns after logging.
	_ = client.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
	buf := make([]byte, 8)
	_, _ = client.Read(buf)
	_ = client.Close()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("handleGRPCConnection did not return")
	}
}

func TestHandleGRPCConnection_AuthRequired(t *testing.T) {
	g := NewGateway(&tunnel.GatewayConfig{}).(*tunnelGateway)
	g.authProvider = &fakeAuth{}
	client, server := net.Pipe()
	t.Cleanup(func() { _ = client.Close() })

	done := make(chan struct{})
	go func() {
		defer close(done)
		g.handleGRPCConnection(server)
	}()

	if _, err := fmt.Fprintf(client, "%ssvc\nbad-auth-line\n", tunnel.GRPCRoutePrefix); err != nil {
		t.Fatal(err)
	}
	_ = client.Close()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("handleGRPCConnection did not return")
	}
}

func TestHandleGRPCConnection_RateLimited(t *testing.T) {
	g := NewGateway(&tunnel.GatewayConfig{}).(*tunnelGateway)
	g.rateLimiter = ratelimit.New(1)
	if !g.rateLimiter.Allow("svc") {
		t.Fatal("seed allow failed")
	}
	// next Allow("svc") should fail

	client, server := net.Pipe()
	t.Cleanup(func() { _ = client.Close() })

	done := make(chan struct{})
	go func() {
		defer close(done)
		g.handleGRPCConnection(server)
	}()

	if _, err := fmt.Fprintf(client, "%ssvc\n", tunnel.GRPCRoutePrefix); err != nil {
		t.Fatal(err)
	}
	_ = client.Close()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("handleGRPCConnection did not return")
	}
}
