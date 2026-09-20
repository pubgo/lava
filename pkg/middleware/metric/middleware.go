package metric

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/pubgo/funk/v2"
	"github.com/pubgo/funk/v2/errors/errcode"
	"github.com/uber-go/tally/v4"

	"github.com/pubgo/lava/v2/core/metrics"
	"github.com/pubgo/lava/v2/pkg/lava"
)

// grpc metric
// ref: https://github.com/grpc-ecosystem/go-grpc-middleware/blob/v2/providers/openmetrics/server_metrics.go
//		https://github.com/grpc-ecosystem/go-grpc-middleware/blob/v2/providers/openmetrics/server_options.go

const (
	rpcTotal          = "lava_rpc_total"
	rpcFailedTotal    = "lava_rpc_failed_total"
	rpcHandlingSecond = "lava_rpc_handling_seconds"
)

// Fast gateway RPCs are the majority, and a histogram that starts at 100ms puts
// every one of them in the same bucket: the number that matters most is the one
// with no resolution.
var requestDurationBucket = tally.DurationBuckets{
	5 * time.Millisecond, 25 * time.Millisecond, 50 * time.Millisecond,
	100 * time.Millisecond, 300 * time.Millisecond,
	1200 * time.Millisecond, 5000 * time.Millisecond, 10000 * time.Millisecond,
}

// side, kind, proto and stream are label values drawn from a request, so this is
// the whole set of series the middleware can produce.
func requestTags(req lava.Request) metrics.Tags {
	return metrics.Tags{
		"side":    side(req),
		"kind":    req.Kind(),
		"service": req.Service(),
		"method":  req.Operation(),
		"stream":  strconv.FormatBool(req.Stream()),
		"proto":   protocol(req.ContentType()),
	}
}

func side(req lava.Request) string {
	if req.Client() {
		return "client"
	}
	return "server"
}

// protocol maps a content type onto a closed set. The gateway takes it from
// caller-supplied request metadata, so the raw value would be an unbounded label.
func protocol(contentType string) string {
	switch {
	case strings.Contains(contentType, "grpc-web"):
		return "grpc-web"
	case strings.HasPrefix(contentType, "application/grpc"):
		return "grpc"
	case strings.Contains(contentType, "json"):
		return "json"
	default:
		return "other"
	}
}

// codeOf labels a failure with the status the caller actually receives. errcode
// reports a bare Go error, which carries no status, as Unknown.
func codeOf(err error) string {
	pb := errcode.ParseError(err)
	if pb == nil {
		return "Unknown"
	}
	return pb.GetStatusCode().String()
}

// recordsRPCs reports whether this request is something the series can describe.
// HTTP traffic is out of scope on purpose: serverhttp and the resty client name
// their operation "VERB /path", and that path is caller-influenced.
func recordsRPCs(req lava.Request) bool {
	return req.Kind() != lava.RequestKindHttp && !strings.Contains(req.Operation(), " ")
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
		if !recordsRPCs(req) {
			return next(ctx, req)
		}

		now := time.Now()
		tagged := m.m.Tagged(requestTags(req))
		tagged.Counter(rpcTotal).Inc(1)

		defer func() {
			if !funk.IsNil(gErr) {
				tagged.Tagged(metrics.Tags{"code": codeOf(gErr)}).Counter(rpcFailedTotal).Inc(1)
			}

			tagged.Histogram(rpcHandlingSecond, requestDurationBucket).RecordDuration(time.Since(now))
		}()

		return next(ctx, req)
	}
}
