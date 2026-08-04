package tunnelgateway

import (
	"context"
	"errors"
	"net"
	"testing"

	"github.com/pubgo/lava/v2/core/tunnel"
)

func closePipe(t *testing.T, c net.Conn) {
	t.Helper()
	t.Cleanup(func() { _ = c.Close() })
}

func TestForward_ServiceNotFound(t *testing.T) {
	g := NewGateway(&tunnel.GatewayConfig{}).(*tunnelGateway)
	client, server := net.Pipe()
	closePipe(t, client)
	closePipe(t, server)

	err := g.Forward(context.Background(), "missing", tunnel.EndpointTypeHTTP, client)
	if !errors.Is(err, tunnel.ErrServiceNotFound) {
		t.Fatalf("err=%v", err)
	}
}

func TestForward_SessionClosed(t *testing.T) {
	g := NewGateway(&tunnel.GatewayConfig{}).(*tunnelGateway)
	g.services["demo"] = &registeredService{
		info:    &tunnel.ServiceInfo{Name: "demo"},
		session: &fakeSession{closed: true},
		agent:   "a1",
	}
	client, server := net.Pipe()
	closePipe(t, client)
	closePipe(t, server)

	err := g.Forward(context.Background(), "demo", tunnel.EndpointTypeGRPC, client)
	if !errors.Is(err, tunnel.ErrSessionClosed) {
		t.Fatalf("err=%v", err)
	}
}

func TestForward_NilSession(t *testing.T) {
	g := NewGateway(&tunnel.GatewayConfig{}).(*tunnelGateway)
	g.services["demo"] = &registeredService{
		info:  &tunnel.ServiceInfo{Name: "demo"},
		agent: "a1",
	}
	client, server := net.Pipe()
	closePipe(t, client)
	closePipe(t, server)

	err := g.Forward(context.Background(), "demo", tunnel.EndpointTypeDebug, client)
	if !errors.Is(err, tunnel.ErrSessionClosed) {
		t.Fatalf("err=%v", err)
	}
}
