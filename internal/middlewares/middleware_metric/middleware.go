package middleware_metric

import (
	"context"
	"time"

	"github.com/pubgo/funk/generic"
	"github.com/pubgo/funk/running"
	"github.com/pubgo/lava/core/lavacontexts"
	"github.com/uber-go/tally/v4"

	"github.com/pubgo/lava/core/metrics"
	"github.com/pubgo/lava/lava"
)

// grpc metric
// ref: https://github.com/grpc-ecosystem/go-grpc-middleware/blob/v2/providers/openmetrics/server_metrics.go
//		https://github.com/grpc-ecosystem/go-grpc-middleware/blob/v2/providers/openmetrics/server_options.go

var requestDurationBucket = tally.DurationBuckets{100 * time.Millisecond, 300 * time.Millisecond, 1200 * time.Millisecond, 5000 * time.Millisecond, 10000 * time.Millisecond}

// Total number of rpc call started on the server.
func grpcServerRpcCallTotal(m metrics.Metric, method string) {
	m.Tagged(metrics.Tags{"method": method}).Counter("grpc_server_rpc_total").Inc(1)
}

// Total number of rpc call failed.
func grpcServerRpcErrTotal(m metrics.Metric, method string) {
	m.Tagged(metrics.Tags{"method": method}).Counter("grpc_server_rpc_failed_total").Inc(1)
}

// Histogram of response latency (seconds) of gRPC that had been application-level handled by the server.
func grpcServerHandlingSecondsCount(m metrics.Metric, method string, val time.Duration) {
	m.Tagged(metrics.Tags{"method": method}).
		Histogram("grpc_server_handling_seconds_count", requestDurationBucket).
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
		now := time.Now()

		grpcServerRpcCallTotal(m.m, req.Operation())

		clientInfo := lavacontexts.GetClientInfo(ctx)
		if !req.Client() && clientInfo != nil {
			m.m.Tagged(metrics.Tags{
				"server_name":   running.Project,
				"server_method": req.Operation(),
				"client_name":   clientInfo.GetName(),
				"client_method": clientInfo.GetPath(),
			}).Counter("grpc_server_info").Inc(1)
		}

		defer func() {
			if !generic.IsNil(gErr) {
				grpcServerRpcErrTotal(m.m, req.Operation())
			}

			grpcServerHandlingSecondsCount(m.m, req.Operation(), time.Since(now))
		}()

		return next(ctx, req)
	}
}
