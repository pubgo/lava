package jaeger

import (
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

	var logs = newLog(cfg.ServiceName)
	var tracer tracing.Tracer
	trace, closer, err := cfg.NewTracer(
		config.Reporter(reporter.NewIoReporter(logs)),
		config.Logger(logs),
		config.Metrics(prometheus.New()),
		config.Sampler(jaeger.NewConstSampler(true)),
	)

	xerror.Panic(err)
	tracer.Tracer = trace
	tracer.Closer = closer

	return tracer, nil
}
