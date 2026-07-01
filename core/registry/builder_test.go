package registry

import (
	"context"
	"testing"

	"github.com/pubgo/lava/v2/core/running"
	"github.com/pubgo/lava/v2/core/service"
)

func TestBuildNodeHasPortAndID(t *testing.T) {
	node := buildNode()
	if node.Port == 0 {
		t.Fatalf("expected non-zero port")
	}
	if node.Address == "" {
		t.Fatalf("expected non-empty address")
	}
	if node.Id == "" {
		t.Fatalf("expected non-empty id")
	}
	if node.GetPort() != node.Port {
		t.Fatalf("GetPort() should match Port field")
	}
}

func TestBuildServiceIncludesRegistryMeta(t *testing.T) {
	reg := &stubRegistry{name: "test-reg"}
	s := buildService(reg, true)
	if s.Name != running.Project() {
		t.Fatalf("unexpected service name %q", s.Name)
	}
	if s.Nodes[0].Metadata["registry"] != "test-reg" {
		t.Fatalf("expected registry metadata")
	}
}

type stubRegistry struct{ name string }

func (s *stubRegistry) String() string { return s.name }
func (s *stubRegistry) Register(_ context.Context, _ *service.Service, _ ...RegOpt) error {
	return nil
}

func (s *stubRegistry) Deregister(_ context.Context, _ *service.Service, _ ...DeregOpt) error {
	return nil
}
