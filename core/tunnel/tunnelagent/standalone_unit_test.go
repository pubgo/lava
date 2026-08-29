package tunnelagent

import (
	"testing"

	"github.com/pubgo/lava/v2/core/tunnel"
)

func TestMergeAgentConfig(t *testing.T) {
	base := &Config{
		GatewayAddr:       ":9000",
		Transport:         tunnel.TransportYamux,
		ServiceName:       "old",
		HeartbeatInterval: 5,
	}
	mergeAgentConfig(base, &Config{
		GatewayAddr:       ":9001",
		ServiceName:       "new",
		HeartbeatInterval: 10,
		Endpoints: []tunnel.EndpointConfig{
			{Type: "http", LocalAddr: ":8080"},
		},
		TLS: tunnel.TLSConfig{Enabled: true, CertFile: "c.pem"},
	})
	if base.GatewayAddr != ":9001" {
		t.Fatalf("gateway=%q", base.GatewayAddr)
	}
	if base.ServiceName != "new" {
		t.Fatalf("service=%q", base.ServiceName)
	}
	if base.HeartbeatInterval != 10 {
		t.Fatalf("heartbeat=%d", base.HeartbeatInterval)
	}
	if len(base.Endpoints) != 1 || base.Endpoints[0].LocalAddr != ":8080" {
		t.Fatalf("endpoints=%v", base.Endpoints)
	}
	if !base.TLS.Enabled || base.TLS.CertFile != "c.pem" {
		t.Fatalf("tls=%+v", base.TLS)
	}
	if base.Transport != tunnel.TransportYamux {
		t.Fatalf("transport should keep base value, got %q", base.Transport)
	}
}
