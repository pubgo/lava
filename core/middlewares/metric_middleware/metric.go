package metric_middleware

import (
	"context"
	"time"

	"github.com/pubgo/funk/metric"
	"github.com/pubgo/lava/lava"
	"github.com/uber-go/tally/v4"
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

func Middleware(m metric.Metric) lava.Middleware {
	return func(next lava.HandlerFunc) lava.HandlerFunc {
		return func(ctx context.Context, req lava.Request) (lava.Response, error) {
			grpcServerRpcCallTotal(m, req.Operation())

			var s = time.Now()
			defer func() {
				grpcServerHandlingSecondsCount(m, req.Operation(), time.Since(s))
			}()

			resp, err := next(ctx, req)
			if err != nil {
				grpcServerRpcErrTotal(m, req.Operation())
			}

			return resp, nil
		}
	}
}
