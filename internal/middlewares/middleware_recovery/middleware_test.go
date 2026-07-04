package middleware_recovery_test

import (
	"context"
	"testing"

	"github.com/pubgo/lava/v2/internal/middlewares/middleware_recovery"
	"github.com/pubgo/lava/v2/lava"
)

type stubRequest struct{}

func (stubRequest) Kind() string                 { return "test" }
func (stubRequest) Client() bool                 { return false }
func (stubRequest) Header() *lava.RequestHeader  { return &lava.RequestHeader{} }
func (stubRequest) Payload() any                 { return nil }
func (stubRequest) ContentType() string          { return "" }
func (stubRequest) Service() string              { return "" }
func (stubRequest) Operation() string            { return "" }
func (stubRequest) Endpoint() string             { return "" }
func (stubRequest) Stream() bool                 { return false }

type stubResponse struct{}

func (stubResponse) Header() *lava.ResponseHeader { return &lava.ResponseHeader{} }
func (stubResponse) Payload() any                 { return nil }
func (stubResponse) Stream() bool                 { return false }

func TestRecoveryMiddlewareCatchesPanic(t *testing.T) {
	t.Parallel()
	mw := middleware_recovery.New()
	handler := mw.Middleware(func(ctx context.Context, req lava.Request) (lava.Response, error) {
		panic("boom")
	})
	_, err := handler(context.Background(), stubRequest{})
	if err == nil {
		t.Fatal("expected error from panic recovery")
	}
}

func TestRecoveryMiddlewarePassesThrough(t *testing.T) {
	t.Parallel()
	mw := middleware_recovery.New()
	handler := mw.Middleware(func(ctx context.Context, req lava.Request) (lava.Response, error) {
		return stubResponse{}, nil
	})
	_, err := handler(context.Background(), stubRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
