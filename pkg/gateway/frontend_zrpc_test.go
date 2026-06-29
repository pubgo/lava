package gateway

import "testing"

func TestZrpcSubject(t *testing.T) {
	got := ZrpcSubject("/grpcweb.example.v1.GreeterService/SayHello", "")
	if got != "svc.grpcweb.example.v1.GreeterService/SayHello" {
		t.Fatalf("unexpected subject: %s", got)
	}

	got = ZrpcSubject("/pkg.v1.Service/Method", "custom.")
	if got != "custom.pkg.v1.Service/Method" {
		t.Fatalf("unexpected subject with prefix: %s", got)
	}
}

func TestMux_RegisterZrpcRequiresQueue(t *testing.T) {
	mux := NewMux()
	if err := mux.RegisterZrpc(nil, ZrpcConfig{}); err == nil {
		t.Fatal("expected error for nil server")
	}
}
