package metric_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	tally "github.com/uber-go/tally/v4"

	"github.com/pubgo/lava/v2/pkg/lava"
	"github.com/pubgo/lava/v2/pkg/httputil"
	"github.com/pubgo/lava/v2/pkg/middleware/metric"
)

type stubRequest struct {
	kind      string
	operation string
}

func (s stubRequest) Kind() string               { return s.kind }
func (s stubRequest) Client() bool               { return false }
func (s stubRequest) Header() lava.RequestHeader { return httputil.NewRequestHeader() }
func (s stubRequest) Payload() any               { return nil }
func (s stubRequest) ContentType() string        { return "" }
func (s stubRequest) Service() string            { return "demo" }
func (s stubRequest) Operation() string          { return s.operation }
func (s stubRequest) Endpoint() string           { return s.operation }
func (s stubRequest) Stream() bool               { return false }

type stubResponse struct{}

func (stubResponse) Header() lava.ResponseHeader { return httputil.NewResponseHeader() }
func (stubResponse) Payload() any                { return nil }
func (stubResponse) Stream() bool                { return false }

func counterValue(snap tally.Snapshot, name string) int64 {
	for key, c := range snap.Counters() {
		if strings.Contains(key, name) {
			return c.Value()
		}
	}
	return 0
}

func TestMetricMiddlewareSkipsHTTP(t *testing.T) {
	t.Parallel()

	scope := tally.NewTestScope("test", nil)
	mw := metric.New(scope)
	handler := mw.Middleware(func(ctx context.Context, req lava.Request) (lava.Response, error) {
		return stubResponse{}, nil
	})

	if _, err := handler(context.Background(), stubRequest{kind: "http", operation: "GET /"}); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if len(scope.Snapshot().Counters()) != 0 {
		t.Fatalf("expected no counters for http request, got %+v", scope.Snapshot().Counters())
	}
}

func TestMetricMiddlewareRecordsGRPCCall(t *testing.T) {
	t.Parallel()

	scope := tally.NewTestScope("test", nil)
	mw := metric.New(scope)
	handler := mw.Middleware(func(ctx context.Context, req lava.Request) (lava.Response, error) {
		return stubResponse{}, nil
	})

	const op = "/demo.Demo/Call"
	if _, err := handler(context.Background(), stubRequest{kind: "grpc", operation: op}); err != nil {
		t.Fatalf("handler: %v", err)
	}

	if counterValue(scope.Snapshot(), "grpc_server_rpc_total") != 1 {
		t.Fatalf("rpc total counter missing, counters=%+v", scope.Snapshot().Counters())
	}
}

func TestMetricMiddlewareRecordsError(t *testing.T) {
	t.Parallel()

	scope := tally.NewTestScope("test", nil)
	mw := metric.New(scope)
	handler := mw.Middleware(func(ctx context.Context, req lava.Request) (lava.Response, error) {
		return nil, errors.New("fail")
	})

	const op = "/demo.Demo/Call"
	if _, err := handler(context.Background(), stubRequest{kind: "grpc", operation: op}); err == nil {
		t.Fatal("expected error")
	}

	if counterValue(scope.Snapshot(), "grpc_server_rpc_failed_total") != 1 {
		t.Fatalf("failed counter missing, counters=%+v", scope.Snapshot().Counters())
	}
}
