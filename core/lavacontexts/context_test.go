package lavacontexts

import (
	"context"
	"testing"

	"github.com/pubgo/lava/v2/pkg/httputil"
	"github.com/pubgo/lava/v2/pkg/proto/lavapbv1"
)

func TestReqID(t *testing.T) {
	ctx := context.Background()

	if got := GetReqID(ctx); got != "" {
		t.Fatalf("empty context should return empty req id, got %q", got)
	}

	ctx = CreateCtxWithReqID(ctx, "req-123")
	if got := GetReqID(ctx); got != "req-123" {
		t.Fatalf("expected req-123, got %q", got)
	}
}

func TestClientServerInfo(t *testing.T) {
	ctx := context.Background()

	if got := GetClientInfo(ctx); got != nil {
		t.Fatalf("empty context should return nil client info, got %v", got)
	}
	if got := GetServerInfo(ctx); got != nil {
		t.Fatalf("empty context should return nil server info, got %v", got)
	}

	client := &lavapbv1.ServiceInfo{Name: "client-svc"}
	server := &lavapbv1.ServiceInfo{Name: "server-svc"}
	ctx = CreateCtxWithClientInfo(ctx, client)
	ctx = CreateCtxWithServerInfo(ctx, server)

	if got := GetClientInfo(ctx); got == nil || got.Name != "client-svc" {
		t.Fatalf("expected client-svc, got %v", got)
	}
	if got := GetServerInfo(ctx); got == nil || got.Name != "server-svc" {
		t.Fatalf("expected server-svc, got %v", got)
	}
}

// TestHeadersMissingDoesNotPanic 锁定回归：context 中无 header 时必须返回 nil 而非 panic。
func TestHeadersMissingDoesNotPanic(t *testing.T) {
	ctx := context.Background()

	if got := ReqHeader(ctx); got != nil {
		t.Fatalf("empty context should return nil request header, got %v", got)
	}
	if got := RspHeader(ctx); got != nil {
		t.Fatalf("empty context should return nil response header, got %v", got)
	}
}

func TestHeadersRoundTrip(t *testing.T) {
	ctx := context.Background()

	reqH := httputil.NewRequestHeader()
	rspH := httputil.NewResponseHeader()
	ctx = CreateReqHeader(ctx, reqH)
	ctx = CreateRspHeader(ctx, rspH)

	if got := ReqHeader(ctx); got != reqH {
		t.Fatalf("request header round-trip failed")
	}
	if got := RspHeader(ctx); got != rspH {
		t.Fatalf("response header round-trip failed")
	}
}
