package gatewayserver

import (
	"context"
	"testing"

	"github.com/pubgo/lava/v2/pkg/httputil"
	"github.com/pubgo/lava/v2/pkg/lava"
)

func TestServiceImplString(t *testing.T) {
	t.Parallel()

	s := &serviceImpl{}
	if s.String() != "gateway-server" {
		t.Fatalf("String() = %q, want gateway-server", s.String())
	}

	named := &serviceImpl{name: "custom-gateway"}
	if named.String() != "custom-gateway" {
		t.Fatalf("String() = %q, want custom-gateway", named.String())
	}
}

func TestGetIncomingMetadataEmptyContext(t *testing.T) {
	t.Parallel()

	md := getIncomingMetadata(context.Background())
	if md == nil {
		t.Fatal("expected non-nil metadata map")
	}
	if len(md) != 0 {
		t.Fatalf("expected empty metadata, got %v", md)
	}
}

func TestRPCRequestImplementsLavaRequest(t *testing.T) {
	t.Parallel()

	var _ lava.Request = (*rpcRequest)(nil)

	req := &rpcRequest{
		service:     "echo.Echo",
		method:      "/echo.Echo/Ping",
		url:         "/api/echo",
		contentType: defaultContentType,
		header:      httputil.NewRequestHeader(),
		payload:     struct{}{},
	}
	if req.Kind() != lava.RequestKindGrpc {
		t.Fatalf("Kind() = %q", req.Kind())
	}
	if req.Client() {
		t.Fatal("expected server-side request")
	}
	if req.Service() != "echo.Echo" || req.Operation() != "/echo.Echo/Ping" {
		t.Fatalf("service/method = %q / %q", req.Service(), req.Operation())
	}
}

func TestRPCResponseImplementsLavaResponse(t *testing.T) {
	t.Parallel()

	var _ lava.Response = (*rpcResponse)(nil)

	rsp := &rpcResponse{
		header: httputil.NewResponseHeader(),
		dt:     "ok",
	}
	if rsp.Stream() {
		t.Fatal("expected non-stream response")
	}
	if rsp.Payload() != "ok" {
		t.Fatalf("Payload() = %v", rsp.Payload())
	}
}
