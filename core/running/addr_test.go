package running

import "testing"

func TestHttpListenAddr(t *testing.T) {
	t.Setenv("HTTP_ADDR", "")
	orig := HttpPort.String()
	t.Cleanup(func() { _ = HttpPort.Set(orig) })
	_ = HttpPort.Set("9090")

	if got := HttpListenAddr(); got != ":9090" {
		t.Fatalf("default=%q want :9090", got)
	}

	t.Setenv("HTTP_ADDR", "127.0.0.1:8080")
	if got := HttpListenAddr(); got != "127.0.0.1:8080" {
		t.Fatalf("HTTP_ADDR=%q", got)
	}
}

func TestDebugListenAddr(t *testing.T) {
	t.Setenv("DEBUG_ADDR", "")
	t.Setenv("DEBUG_PORT", "")
	if got := DebugListenAddr(); got != ":6060" {
		t.Fatalf("default=%q want :6060", got)
	}

	t.Setenv("DEBUG_PORT", "7070")
	if got := DebugListenAddr(); got != ":7070" {
		t.Fatalf("DEBUG_PORT=%q", got)
	}

	t.Setenv("DEBUG_ADDR", "127.0.0.1:6060")
	if got := DebugListenAddr(); got != "127.0.0.1:6060" {
		t.Fatalf("DEBUG_ADDR=%q", got)
	}
}
