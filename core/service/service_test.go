package service

import "testing"

func TestNodeGetPortExplicit(t *testing.T) {
	n := Node{Port: 8080, Address: "ignored:9999"}
	if got := n.GetPort(); got != 8080 {
		t.Fatalf("expected explicit port 8080, got %d", got)
	}
}

func TestNodeGetPortFromAddress(t *testing.T) {
	n := Node{Address: "127.0.0.1:9090"}
	if got := n.GetPort(); got != 9090 {
		t.Fatalf("expected port 9090 from address, got %d", got)
	}
}

func TestNodeGetPortIPv6(t *testing.T) {
	n := Node{Address: "[::1]:50051"}
	if got := n.GetPort(); got != 50051 {
		t.Fatalf("expected port 50051 from ipv6 address, got %d", got)
	}
}

func TestNodeGetPortMissing(t *testing.T) {
	n := Node{Address: "no-port-here"}
	if got := n.GetPort(); got != 0 {
		t.Fatalf("expected 0 when no port available, got %d", got)
	}
}
