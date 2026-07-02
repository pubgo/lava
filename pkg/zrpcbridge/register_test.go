package zrpcbridge

import (
	"testing"

	"github.com/pubgo/lava/v2/pkg/gateway"
)

func TestSubject(t *testing.T) {
	got := Subject("/grpcweb.example.v1.GreeterService/SayHello", "")
	if got != "svc.grpcweb.example.v1.GreeterService/SayHello" {
		t.Fatalf("unexpected subject: %s", got)
	}

	got = Subject("/pkg.v1.Service/Method", "custom.")
	if got != "custom.pkg.v1.Service/Method" {
		t.Fatalf("unexpected subject with prefix: %s", got)
	}
}

func TestRegisterMuxRequiresQueue(t *testing.T) {
	mux := gateway.NewMux()
	if err := RegisterMux(nil, mux, Config{}); err == nil {
		t.Fatal("expected error for nil server")
	}
}
