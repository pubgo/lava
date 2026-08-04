package tunnelgateway

import (
	"testing"

	"github.com/pubgo/lava/v2/core/tunnel"
)

func TestBuilder_BuildRequiresListenAddr(t *testing.T) {
	_, err := NewBuilder().WithListenAddr("").Build()
	if err == nil {
		t.Fatal("expected error for empty listen addr")
	}

	gw, err := NewBuilder().WithListenAddr(":9000").Build()
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if gw.Config().Transport != TransportYamux {
		t.Fatalf("default transport=%q, want yamux", gw.Config().Transport)
	}
	if gw.Inner() == nil {
		t.Fatal("Inner() is nil")
	}
}

func TestBuilder_WithTransportPreserved(t *testing.T) {
	gw, err := NewBuilder().
		WithListenAddr(":9001").
		WithTransport(TransportQUIC).
		WithHTTPPort(8080).
		Build()
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	cfg := gw.Config()
	if cfg.Transport != TransportQUIC {
		t.Fatalf("transport=%q", cfg.Transport)
	}
	if cfg.HTTPPort != 8080 {
		t.Fatalf("http port=%d", cfg.HTTPPort)
	}
}

func TestEndpointTypeToMessageType(t *testing.T) {
	g := NewGateway(&tunnel.GatewayConfig{}).(*tunnelGateway)
	tests := []struct {
		et   tunnel.EndpointType
		want tunnel.MessageType
	}{
		{tunnel.EndpointTypeHTTP, tunnel.MessageTypeHTTPRequest},
		{tunnel.EndpointTypeGRPC, tunnel.MessageTypeGRPCRequest},
		{tunnel.EndpointTypeDebug, tunnel.MessageTypeDebugRequest},
		{tunnel.EndpointType("unknown"), tunnel.MessageTypeHTTPRequest},
	}
	for _, tt := range tests {
		t.Run(string(tt.et), func(t *testing.T) {
			if got := g.endpointTypeToMessageType(tt.et); got != tt.want {
				t.Fatalf("got=%v want=%v", got, tt.want)
			}
		})
	}
}

func TestP2PRateLimitFromConfig_ZeroUsesDefault(t *testing.T) {
	for _, cfg := range []*tunnel.GatewayConfig{
		nil,
		{P2PSignalRateLimit: 0, P2PRegisterRateLimit: 0},
		{P2PSignalRateLimit: -1, P2PRegisterRateLimit: -5},
	} {
		if got := p2pSignalRateLimit(cfg); got != 60 {
			t.Fatalf("signal default for %+v = %d", cfg, got)
		}
		if got := p2pRegisterRateLimit(cfg); got != 10 {
			t.Fatalf("register default for %+v = %d", cfg, got)
		}
	}
}
