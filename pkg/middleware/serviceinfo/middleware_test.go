package serviceinfo_test

import (
	"context"
	"testing"

	"github.com/pubgo/lava/v2/lava"
	"github.com/pubgo/lava/v2/pkg/grpcutil"
	"github.com/pubgo/lava/v2/pkg/httputil"
	"github.com/pubgo/lava/v2/pkg/lavacontexts"
	"github.com/pubgo/lava/v2/pkg/middleware/serviceinfo"
)

type stubRequest struct {
	header lava.RequestHeader
	client bool
}

func (s stubRequest) Kind() string               { return "grpc" }
func (s stubRequest) Client() bool               { return s.client }
func (s stubRequest) Header() lava.RequestHeader { return s.header }
func (s stubRequest) Payload() any               { return nil }
func (s stubRequest) ContentType() string        { return "" }
func (s stubRequest) Service() string            { return "demo" }
func (s stubRequest) Operation() string          { return "/demo.Demo/Call" }
func (s stubRequest) Endpoint() string           { return "/demo.Demo/Call" }
func (s stubRequest) Stream() bool               { return false }

type stubResponse struct {
	header lava.ResponseHeader
}

func (s stubResponse) Header() lava.ResponseHeader { return s.header }
func (s stubResponse) Payload() any                { return nil }
func (s stubResponse) Stream() bool                { return false }

func TestServiceInfoClientPopulatesServerContext(t *testing.T) {
	t.Parallel()

	mw := serviceinfo.New()
	handler := mw.Middleware(func(ctx context.Context, req lava.Request) (lava.Response, error) {
		info := lavacontexts.GetServerInfo(ctx)
		if info == nil || info.GetPath() != "/demo.Demo/Call" {
			t.Fatalf("server info = %+v", info)
		}
		return stubResponse{header: httputil.NewResponseHeader()}, nil
	})

	req := stubRequest{header: httputil.NewRequestHeader(), client: true}
	if _, err := handler(context.Background(), req); err != nil {
		t.Fatalf("handler: %v", err)
	}
}

func TestServiceInfoServerReadsClientInfoFromHeaders(t *testing.T) {
	t.Parallel()

	mw := serviceinfo.New()
	header := httputil.NewRequestHeader()
	header.Set(grpcutil.ClientNameKey, "remote-svc")
	header.Set(grpcutil.ClientPathKey, "/remote.Op")

	handler := mw.Middleware(func(ctx context.Context, req lava.Request) (lava.Response, error) {
		info := lavacontexts.GetClientInfo(ctx)
		if info == nil || info.GetName() != "remote-svc" || info.GetPath() != "/remote.Op" {
			t.Fatalf("client info = %+v", info)
		}
		return stubResponse{header: httputil.NewResponseHeader()}, nil
	})

	req := stubRequest{header: header, client: false}
	if _, err := handler(context.Background(), req); err != nil {
		t.Fatalf("handler: %v", err)
	}
}

func TestServiceInfoSetsResponseRequestID(t *testing.T) {
	t.Parallel()

	const wantID = "fixed-req-id"
	ctx := lavacontexts.CreateCtxWithReqID(context.Background(), wantID)

	mw := serviceinfo.New()
	rspHeader := httputil.NewResponseHeader()
	handler := mw.Middleware(func(ctx context.Context, req lava.Request) (lava.Response, error) {
		return stubResponse{header: rspHeader}, nil
	})

	req := stubRequest{header: httputil.NewRequestHeader(), client: false}
	if _, err := handler(ctx, req); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if got := string(rspHeader.Peek(httputil.HeaderXRequestID)); got != wantID {
		t.Fatalf("response request id = %q, want %q", got, wantID)
	}
}
