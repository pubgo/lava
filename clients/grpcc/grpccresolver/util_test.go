package grpccresolver

import (
	"strings"
	"testing"
)

func TestBuildDirectTarget(t *testing.T) {
	t.Parallel()

	got := BuildDirectTarget("mysvc", "127.0.0.1:8080", "127.0.0.1:8081")
	if !strings.HasPrefix(got, DirectScheme+"://") {
		t.Fatalf("unexpected scheme prefix: %q", got)
	}
	if !strings.Contains(got, "name=mysvc") {
		t.Fatalf("missing name param: %q", got)
	}
	if !strings.Contains(got, "127.0.0.1:8080") || !strings.Contains(got, "127.0.0.1:8081") {
		t.Fatalf("missing endpoints: %q", got)
	}
}

func TestBuildDiscoveryTarget(t *testing.T) {
	t.Parallel()

	got := BuildDiscoveryTarget("test-service")
	want := DiscoveryScheme + "://test-service"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestGetServiceUniqueId(t *testing.T) {
	t.Parallel()

	if got := getServiceUniqueId("svc", 3); got != "svc-3" {
		t.Fatalf("got %q", got)
	}
}

func TestNewAddr(t *testing.T) {
	t.Parallel()

	addr := newAddr("10.0.0.1:443", "api.example.com")
	if addr.Addr != "10.0.0.1:443" || addr.ServerName != "api.example.com" {
		t.Fatalf("addr=%+v", addr)
	}
}
