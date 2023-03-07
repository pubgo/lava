package middleware_metric

import (
	"context"
	"time"

	"github.com/pubgo/funk/strutil"
	"github.com/pubgo/funk/version"
	"github.com/rs/xid"
	"github.com/uber-go/tally/v4"

	"github.com/pubgo/lava"
	"github.com/pubgo/lava/core/metric"
	"github.com/pubgo/lava/pkg/httputil"
)

// grpc metric
// ref: https://github.com/grpc-ecosystem/go-grpc-middleware/blob/v2/providers/openmetrics/server_metrics.go
//		https://github.com/grpc-ecosystem/go-grpc-middleware/blob/v2/providers/openmetrics/server_options.go

var requestDurationBucket = tally.DurationBuckets{100 * time.Millisecond, 300 * time.Millisecond, 1200 * time.Millisecond, 5000 * time.Millisecond, 10000 * time.Millisecond}

// Total number of rpc call started on the server.
func grpcServerRpcCallTotal(m metric.Metric, method string) {
	m.Tagged(metric.Tags{"method": method}).Counter("grpc_server_rpc_total").Inc(1)
}

// Total number of rpc call failed.
func grpcServerRpcErrTotal(m metric.Metric, method string) {
	m.Tagged(metric.Tags{"method": method}).Counter("grpc_server_rpc_failed_total").Inc(1)
}

// Histogram of response latency (seconds) of gRPC that had been application-level handled by the server.
func grpcServerHandlingSecondsCount(m metric.Metric, method string, val time.Duration) {
	m.Tagged(metric.Tags{"method": method}).
		Histogram("grpc_server_handling_seconds_count", requestDurationBucket).
		RecordValue(float64(val.Milliseconds()))
}

func New(m metric.Metric) lava.Middleware {
	return func(next lava.HandlerFunc) lava.HandlerFunc {
		return func(ctx context.Context, req lava.Request) (rsp lava.Response, gErr error) {
			var now = time.Now()

			reqId := strutil.FirstFnNotEmpty(
				func() string { return lava.GetReqID(ctx) },
				func() string { return string(req.Header().Peek(httputil.HeaderXRequestID)) },
				func() string { return xid.New().String() },
			)

			grpcServerRpcCallTotal(m, req.Operation())

			req.Header().Set(httputil.HeaderXRequestID, reqId)

			ctx = lava.CreateCtxWithReqID(ctx, reqId)
			rsp, gErr = next(ctx, req)
			if gErr != nil {
				grpcServerRpcErrTotal(m, req.Operation())
			}

			rsp.Header().Set(httputil.HeaderXRequestID, reqId)
			rsp.Header().Set(httputil.HeaderXRequestVersion, version.Version())
			grpcServerHandlingSecondsCount(m, req.Operation(), time.Since(now))

			return rsp, nil
		}
	}
}
