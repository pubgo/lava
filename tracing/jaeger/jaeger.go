package jaeger

import (
	"net/http"
	"strings"

	"github.com/pubgo/lug/runenv"
	"github.com/pubgo/lug/tracing"
	"github.com/pubgo/lug/tracing/jaeger/reporter"

	"github.com/opentracing/opentracing-go"
	"github.com/pubgo/x/merge"
	"github.com/pubgo/xerror"
	"github.com/uber/jaeger-client-go"
	"github.com/uber/jaeger-client-go/config"
	"github.com/uber/jaeger-lib/metrics/prometheus"
)

var _ = jaeger.NewNullReporter()

const (
	phpRequestTraceID      = "x-w-traceid"
	phpRequestSpanID       = "x-w-spanid"
	phpRequestParentSpanID = "x-w-parentspanid"
	phpRequestSampleID     = "x-w-sampled"
)

func init() {
	tracing.GetTraceId = func(span opentracing.SpanContext) string {
		var ctx, ok = span.(jaeger.SpanContext)
		if !ok {
			return ""
		}

		var traceID = ctx.TraceID()
		if !traceID.IsValid() {
			return ""
		}

		return traceID.String()
	}

	xerror.Exit(tracing.Register(Name, func(cfgMap map[string]interface{}) (tracing.Tracer, error) {
		var cfg = GetDefaultCfg()
		cfg.ServiceName = runenv.Project

		xerror.Panic(merge.MapStruct(&cfg, cfgMap))
		return New(cfg)
	}))
}

func New(cfg *Cfg) (tracing.Tracer, error) {
	cfg.Disabled = false
	if cfg.ServiceName == "" {
		cfg.ServiceName = runenv.Project
	}

	var logs = newLog("tracing")
	var tracer tracing.Tracer
	trace, _, err := cfg.NewTracer(
		config.Reporter(reporter.NewIoReporter(logs, cfg.BatchSize)),
		config.Logger(logs),
		config.Metrics(prometheus.New()),
		config.Sampler(jaeger.NewConstSampler(true)),
	)
	xerror.Exit(err)
	tracer.Tracer = trace

	return tracer, nil
}

func spanFromPHPRequest(req *http.Request) (span jaeger.SpanContext, err error) {
	defer xerror.RespErr(&err)

	if req == nil {
		return span, xerror.Fmt("context is nil")
	}

	var sampleIDStr = strings.Join(req.Header.Values(phpRequestSampleID), ",")
	var traceIDStr = strings.Join(req.Header.Values(phpRequestTraceID), ",")
	traceID, err := jaeger.TraceIDFromString(traceIDStr)
	xerror.Panic(err)

	var spanIDStr = strings.Join(req.Header.Values(phpRequestSpanID), ",")
	spanID, err := jaeger.SpanIDFromString(spanIDStr)
	xerror.Panic(err)

	var pSpanIDStr = strings.Join(req.Header.Values(phpRequestParentSpanID), ",")
	pSpanID, err := jaeger.SpanIDFromString(pSpanIDStr)
	xerror.Panic(err)

	return jaeger.NewSpanContext(traceID, spanID, pSpanID, sampleIDStr == "", nil), nil
}
