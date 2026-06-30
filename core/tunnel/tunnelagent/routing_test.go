package tunnelagent

import (
	"testing"

	"github.com/pubgo/lava/v2/core/tunnel"
)

func TestFindEndpointByServiceID(t *testing.T) {
	services := map[string]*tunnel.ServiceInfo{
		"svc-a": {
			ID:   "id-a",
			Name: "svc-a",
			Endpoints: []tunnel.Endpoint{
				{Type: tunnel.EndpointTypeHTTP, Address: ":8081"},
			},
		},
		"svc-b": {
			ID:   "id-b",
			Name: "svc-b",
			Endpoints: []tunnel.Endpoint{
				{Type: tunnel.EndpointTypeHTTP, Address: ":8082"},
			},
		},
	}

	ep := findEndpoint(services, tunnel.RequestMeta{ServiceID: "id-b"}, tunnel.EndpointTypeHTTP)
	if ep == nil || ep.Address != ":8082" {
		t.Fatalf("expected svc-b endpoint, got %v", ep)
	}

	if ep := findEndpoint(services, tunnel.RequestMeta{ServiceID: "missing"}, tunnel.EndpointTypeHTTP); ep != nil {
		t.Fatalf("expected nil for unknown service id, got %v", ep)
	}
}

func TestFindEndpointFallbackFirst(t *testing.T) {
	services := map[string]*tunnel.ServiceInfo{
		"only": {
			Name: "only",
			Endpoints: []tunnel.Endpoint{
				{Type: tunnel.EndpointTypeGRPC, Address: ":50051"},
			},
		},
	}

	ep := findEndpoint(services, tunnel.RequestMeta{}, tunnel.EndpointTypeGRPC)
	if ep == nil || ep.Address != ":50051" {
		t.Fatalf("expected grpc endpoint, got %v", ep)
	}
}

func TestNormalizeAddress(t *testing.T) {
	if got := normalizeAddress(":8080"); got != "127.0.0.1:8080" {
		t.Fatalf("expected 127.0.0.1:8080, got %q", got)
	}
	if got := normalizeAddress("10.0.0.1:8080"); got != "10.0.0.1:8080" {
		t.Fatalf("expected unchanged address, got %q", got)
	}
}
