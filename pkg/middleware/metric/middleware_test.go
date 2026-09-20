package metric_test

import (
	"context"
	"errors"
	"slices"
	"strings"
	"testing"
	"time"

	tally "github.com/uber-go/tally/v4"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/pubgo/lava/v2/pkg/httputil"
	"github.com/pubgo/lava/v2/pkg/lava"
	"github.com/pubgo/lava/v2/pkg/middleware/metric"
)

type stubRequest struct {
	kind        string
	operation   string
	contentType string
	client      bool
	streamed    bool
}

func (s stubRequest) Kind() string               { return s.kind }
func (s stubRequest) Client() bool               { return s.client }
func (s stubRequest) Header() lava.RequestHeader { return httputil.NewRequestHeader() }
func (s stubRequest) Payload() any               { return nil }
func (s stubRequest) ContentType() string        { return s.contentType }
func (s stubRequest) Service() string            { return "demo" }
func (s stubRequest) Operation() string          { return s.operation }
func (s stubRequest) Endpoint() string           { return s.operation }
func (s stubRequest) Stream() bool               { return s.streamed }

type stubResponse struct{}

func (stubResponse) Header() lava.ResponseHeader { return httputil.NewResponseHeader() }
func (stubResponse) Payload() any                { return nil }
func (stubResponse) Stream() bool                { return false }

// seriesTags returns the label set of every recorded series named want.
func seriesTags(snap tally.Snapshot, want string) []map[string]string {
	var out []map[string]string
	for _, c := range snap.Counters() {
		if strings.HasSuffix(c.Name(), want) {
			out = append(out, c.Tags())
		}
	}
	for _, h := range snap.Histograms() {
		if strings.HasSuffix(h.Name(), want) {
			out = append(out, h.Tags())
		}
	}
	return out
}

// histogramNames returns every recorded histogram series name.
func histogramNames(snap tally.Snapshot) []string {
	var out []string
	for _, h := range snap.Histograms() {
		out = append(out, h.Name())
	}
	return out
}

// assertLabeled fails unless a series exists carrying at least the wanted labels.
func assertLabeled(t *testing.T, snap tally.Snapshot, want string, labels map[string]string) {
	t.Helper()

	for _, tags := range seriesTags(snap, want) {
		matched := true
		for k, v := range labels {
			if tags[k] != v {
				matched = false
				break
			}
		}
		if matched {
			return
		}
	}

	t.Fatalf("no series named %s has labels %v, got %v", want, labels, seriesTags(snap, want))
}

func counterValue(snap tally.Snapshot, name string) int64 {
	for _, c := range snap.Counters() {
		if strings.HasSuffix(c.Name(), name) {
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

	if counterValue(scope.Snapshot(), "lava_rpc_total") != 1 {
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

	if counterValue(scope.Snapshot(), "lava_rpc_failed_total") != 1 {
		t.Fatalf("failed counter missing, counters=%+v", scope.Snapshot().Counters())
	}
}

func okResponse(context.Context, lava.Request) (lava.Response, error) {
	return stubResponse{}, nil
}

func failingResponse(err error) lava.HandlerFunc {
	return func(context.Context, lava.Request) (lava.Response, error) {
		return nil, err
	}
}

func TestMetricMiddlewareLabelsEverySeries(t *testing.T) {
	t.Parallel()

	scope := tally.NewTestScope("test", nil)
	handler := metric.New(scope).Middleware(okResponse)

	const op = "/demo.Demo/Call"
	if _, err := handler(context.Background(), stubRequest{
		kind:        "grpc",
		operation:   op,
		contentType: "application/grpc-web+proto",
		streamed:    true,
	}); err != nil {
		t.Fatalf("handler: %v", err)
	}

	// The gateway serves one operation over several protocols, so a bare method
	// label cannot tell a slow grpc-web browser client from a fast json one.
	want := map[string]string{
		"side":    "server",
		"kind":    "grpc",
		"service": "demo",
		"method":  op,
		"stream":  "true",
		"proto":   "grpc-web",
	}
	snap := scope.Snapshot()
	assertLabeled(t, snap, "lava_rpc_total", want)
	assertLabeled(t, snap, "lava_rpc_handling_seconds", want)
}

func TestMetricMiddlewareSeparatesServerFromClient(t *testing.T) {
	t.Parallel()

	scope := tally.NewTestScope("test", nil)
	handler := metric.New(scope).Middleware(okResponse)

	const op = "/demo.Demo/Call"
	if _, err := handler(context.Background(), stubRequest{kind: "grpc", operation: op}); err != nil {
		t.Fatalf("server handler: %v", err)
	}
	if _, err := handler(context.Background(), stubRequest{kind: "grpcc", operation: op, client: true}); err != nil {
		t.Fatalf("client handler: %v", err)
	}

	// Servers and clients install this same middleware; without a side label a
	// client's failed call is indistinguishable from a server rejecting one.
	snap := scope.Snapshot()
	assertLabeled(t, snap, "lava_rpc_total", map[string]string{"side": "server", "kind": "grpc"})
	assertLabeled(t, snap, "lava_rpc_total", map[string]string{"side": "client", "kind": "grpcc"})
}

func TestMetricMiddlewareNormalizesProtocolLabel(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		contentType string
		want        string
	}{
		{"application/grpc-web+proto", "grpc-web"},
		{"application/grpc", "grpc"},
		{"application/json", "json"},
		{"", "other"},
		{"application/x-lava-custom", "other"},
	} {
		t.Run(tc.want+" for "+tc.contentType, func(t *testing.T) {
			t.Parallel()

			// The gateway reads content type out of caller-supplied metadata, so only
			// a closed set of values may become a label.
			scope := tally.NewTestScope("test", nil)
			handler := metric.New(scope).Middleware(okResponse)
			if _, err := handler(context.Background(), stubRequest{
				kind:        "grpc",
				operation:   "/demo.Demo/Call",
				contentType: tc.contentType,
			}); err != nil {
				t.Fatalf("handler: %v", err)
			}

			assertLabeled(t, scope.Snapshot(), "lava_rpc_total", map[string]string{"proto": tc.want})
		})
	}
}

func TestMetricMiddlewareCodesFailedCalls(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name string
		err  error
		want string
	}{
		{"grpc status", status.Error(codes.NotFound, "no such widget"), "NotFound"},
		{"plain error", errors.New("fail"), "Unknown"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			scope := tally.NewTestScope("test", nil)
			handler := metric.New(scope).Middleware(failingResponse(tc.err))
			if _, err := handler(context.Background(), stubRequest{kind: "grpc", operation: "/demo.Demo/Call"}); err == nil {
				t.Fatal("expected error")
			}

			snap := scope.Snapshot()
			assertLabeled(t, snap, "lava_rpc_failed_total", map[string]string{"code": tc.want})
			for _, tags := range seriesTags(snap, "lava_rpc_total") {
				if _, ok := tags["code"]; ok {
					t.Fatalf("the success counter must not carry a code label: %v", tags)
				}
			}
		})
	}
}

func TestMetricMiddlewareHistogramIsNamedByItsUnit(t *testing.T) {
	t.Parallel()

	scope := tally.NewTestScope("test", nil)
	handler := metric.New(scope).Middleware(okResponse)
	if _, err := handler(context.Background(), stubRequest{kind: "grpc", operation: "/demo.Demo/Call"}); err != nil {
		t.Fatalf("handler: %v", err)
	}

	// tally hands the name straight to prometheus, which appends _bucket, _count and
	// _sum itself, so a name already ending in _count exports seconds_count_count.
	names := histogramNames(scope.Snapshot())
	if len(names) != 1 {
		t.Fatalf("expected one histogram, got %v", names)
	}
	if !strings.HasSuffix(names[0], "lava_rpc_handling_seconds") {
		t.Fatalf("histogram name %q is not lava_rpc_handling_seconds", names[0])
	}
}

func TestMetricMiddlewareHistogramBucketsCoverFastRPCs(t *testing.T) {
	t.Parallel()

	scope := tally.NewTestScope("test", nil)
	handler := metric.New(scope).Middleware(okResponse)
	if _, err := handler(context.Background(), stubRequest{kind: "grpc", operation: "/demo.Demo/Call"}); err != nil {
		t.Fatalf("handler: %v", err)
	}

	var bounds []time.Duration
	for _, h := range scope.Snapshot().Histograms() {
		if !strings.HasSuffix(h.Name(), "lava_rpc_handling_seconds") {
			continue
		}
		for upperBound := range h.Durations() {
			bounds = append(bounds, upperBound)
		}
	}

	// A gateway RPC is usually well under 100ms. With no lower bucket every fast call
	// lands in the same one, and a regression there is invisible.
	if !slices.Contains(bounds, 50*time.Millisecond) {
		t.Fatalf("expected a 50ms bucket, got %v", bounds)
	}
}

func TestMetricMiddlewareSkipsHTTPShapedOperations(t *testing.T) {
	t.Parallel()

	scope := tally.NewTestScope("test", nil)
	handler := metric.New(scope).Middleware(okResponse)

	// resty reports its kind as "resty", so an http-kind check alone would let a
	// "POST /path" operation through as though it were an RPC method name.
	if _, err := handler(context.Background(), stubRequest{
		kind: "resty", operation: "POST /v1/widgets", client: true,
	}); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if len(scope.Snapshot().Counters()) != 0 {
		t.Fatalf("expected no counters for an http-shaped operation, got %+v", scope.Snapshot().Counters())
	}
}
