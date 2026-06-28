package tunnel_test

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/pubgo/lava/v2/core/tunnel"
)

func TestGatewayConfigFromEnv(t *testing.T) {
	t.Setenv("TUNNEL_LISTEN_ADDR", ":9001")
	t.Setenv("TUNNEL_HTTP_PORT", "8888")
	t.Setenv("TUNNEL_P2P_SIGNAL_RATE_LIMIT", "30")

	cfg := tunnel.GatewayConfigFromEnv()
	if cfg.ListenAddr != ":9001" {
		t.Fatalf("listen=%q", cfg.ListenAddr)
	}
	if cfg.HTTPPort != 8888 {
		t.Fatalf("http=%d", cfg.HTTPPort)
	}
	if cfg.P2PSignalRateLimit != 30 {
		t.Fatalf("p2p signal=%d", cfg.P2PSignalRateLimit)
	}
}

func TestAgentConfigFromEnv(t *testing.T) {
	t.Setenv("TUNNEL_GATEWAY_ADDR", "gw:7007")
	t.Setenv("SERVICE_NAME", "worker-a")
	t.Setenv("TUNNEL_AUTH_TOKEN", "secret")

	cfg := tunnel.AgentConfigFromEnv()
	if cfg.GatewayAddr != "gw:7007" {
		t.Fatalf("gateway=%q", cfg.GatewayAddr)
	}
	if cfg.ServiceName != "worker-a" {
		t.Fatalf("service=%q", cfg.ServiceName)
	}
	if cfg.Metadata["auth_token"] != "secret" {
		t.Fatalf("token=%q", cfg.Metadata["auth_token"])
	}
}

func TestGRPCRouteLine(t *testing.T) {
	line := tunnel.GRPCRouteLine("my-svc")
	if line != "TUNNEL my-svc\n" {
		t.Fatalf("line=%q", line)
	}
	name, ok := tunnel.ParseGRPCRouteLine("TUNNEL my-svc")
	if !ok || name != "my-svc" {
		t.Fatalf("parse=%q ok=%v", name, ok)
	}
}

func TestServiceURL(t *testing.T) {
	got := tunnel.ServiceURL("http://127.0.0.1:8080", "api", "/v1/ping")
	want := "http://127.0.0.1:8080/api/v1/ping"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestGRPCContextDialer(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()

	go func() {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		buf := make([]byte, 64)
		n, _ := conn.Read(buf)
		if string(buf[:n]) != tunnel.GRPCRouteLine("echo") {
			return
		}
		_, _ = conn.Write([]byte("ok"))
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	dialer := tunnel.GRPCContextDialer(tunnel.GRPCDialOptions{GatewayAddr: ln.Addr().String()})
	conn, err := dialer(ctx, "echo")
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	buf := make([]byte, 8)
	n, err := conn.Read(buf)
	if err != nil {
		t.Fatal(err)
	}
	if string(buf[:n]) != "ok" {
		t.Fatalf("got %q", buf[:n])
	}
}
