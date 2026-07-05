package accesslog_test

import (
	"context"
	"testing"

	"github.com/pubgo/funk/v2/log"

	"github.com/pubgo/lava/v2/lava"
	"github.com/pubgo/lava/v2/pkg/httputil"
	"github.com/pubgo/lava/v2/pkg/middleware/accesslog"
)

type stubRequest struct {
	header lava.RequestHeader
	client bool
}

func (s stubRequest) Kind() string               { return "grpc" }
func (s stubRequest) Client() bool               { return s.client }
func (s stubRequest) Header() lava.RequestHeader { return s.header }
func (s stubRequest) Payload() any               { return nil }
func (s stubRequest) ContentType() string        { return "application/grpc" }
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

func TestAccessLogSetsLatencyHeaderOnServer(t *testing.T) {
	t.Parallel()

	mw := accesslog.New(log.GetLogger("test"))
	rspHeader := httputil.NewResponseHeader()
	handler := mw.Middleware(func(ctx context.Context, req lava.Request) (lava.Response, error) {
		return stubResponse{header: rspHeader}, nil
	})

	req := stubRequest{header: httputil.NewRequestHeader(), client: false}
	if _, err := handler(context.Background(), req); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if len(rspHeader.Peek("X-Request-Latency")) == 0 {
		t.Fatal("expected X-Request-Latency response header")
	}
}

func TestAccessLogPassesThroughSuccess(t *testing.T) {
	t.Parallel()

	mw := accesslog.New(log.GetLogger("test"))
	handler := mw.Middleware(func(ctx context.Context, req lava.Request) (lava.Response, error) {
		return stubResponse{header: httputil.NewResponseHeader()}, nil
	})

	req := stubRequest{header: httputil.NewRequestHeader(), client: true}
	if _, err := handler(context.Background(), req); err != nil {
		t.Fatalf("handler: %v", err)
	}
}
