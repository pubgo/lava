package metric

import (
	"context"
	"strings"
	"time"

	"github.com/pubgo/funk/v2"
	"github.com/uber-go/tally/v4"

	"github.com/pubgo/lava/v2/core/metrics"
	"github.com/pubgo/lava/v2/pkg/lava"
)

// grpc metric
// ref: https://github.com/grpc-ecosystem/go-grpc-middleware/blob/v2/providers/openmetrics/server_metrics.go
//		https://github.com/grpc-ecosystem/go-grpc-middleware/blob/v2/providers/openmetrics/server_options.go

var requestDurationBucket = tally.DurationBuckets{100 * time.Millisecond, 300 * time.Millisecond, 1200 * time.Millisecond, 5000 * time.Millisecond, 10000 * time.Millisecond}

// Total number of rpc call started on the server.
func gatewayServerRpcCallTotal(m metrics.Metric, method string) {
	m.Tagged(metrics.Tags{"method": method}).Counter("gateway_server_rpc_total").Inc(1)
}

func gatewayServerRpcErrTotal(m metrics.Metric, method string) {
	m.Tagged(metrics.Tags{"method": method}).Counter("gateway_server_rpc_failed_total").Inc(1)
}

func gatewayServerHandlingSecondsCount(m metrics.Metric, method string, val time.Duration) {
	m.Tagged(metrics.Tags{"method": method}).
		Histogram("gateway_server_handling_seconds_count", requestDurationBucket).
		RecordDuration(val)
}

func New(m metrics.Metric) *MetricMiddleware {
	return &MetricMiddleware{m: m}
}

var _ lava.Middleware = (*MetricMiddleware)(nil)

type MetricMiddleware struct {
	m metrics.Metric
}

func (m MetricMiddleware) String() string { return "metric" }

func (m MetricMiddleware) Middleware(next lava.HandlerFunc) lava.HandlerFunc {
	return func(ctx context.Context, req lava.Request) (rsp lava.Response, gErr error) {
		if req.Kind() == "http" || strings.Contains(req.Operation(), " ") {
			return next(ctx, req)
		}

		now := time.Now()
		gatewayServerRpcCallTotal(m.m, req.Operation())

		defer func() {
			if !funk.IsNil(gErr) {
				gatewayServerRpcErrTotal(m.m, req.Operation())
			}

			gatewayServerHandlingSecondsCount(m.m, req.Operation(), time.Since(now))
		}()

		return next(ctx, req)
	}
}
