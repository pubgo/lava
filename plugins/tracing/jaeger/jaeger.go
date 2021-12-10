package jaeger

import (
	"net/http"
	"strings"

	"github.com/opentracing/opentracing-go"
	"github.com/pubgo/xerror"
	"github.com/uber/jaeger-client-go"
	"github.com/uber/jaeger-client-go/config"
	"github.com/uber/jaeger-lib/metrics/prometheus"

	"github.com/pubgo/lava/pkg/merge"
	"github.com/pubgo/lava/plugins/tracing"
	"github.com/pubgo/lava/plugins/tracing/jaeger/reporter"
	"github.com/pubgo/lava/runenv"
	"github.com/pubgo/lava/version"
)

var _ = jaeger.NewNullReporter()

const (
	// molten兼容
	phpRequestTraceID      = "x-w-traceid"
	phpRequestSpanID       = "x-w-spanid"
	phpRequestParentSpanID = "x-w-parentspanid"
	phpRequestSampleID     = "x-w-sampled"
)

func init() {
	xerror.Exit(tracing.Register(Name, func(cfgMap map[string]interface{}) error {
		var cfg = DefaultCfg()
		cfg.ServiceName = runenv.Project
		cfg.Tags = append(cfg.Tags, opentracing.Tag{Key: "version", Value: version.Version})
		return New(merge.MapStruct(cfg, cfgMap).(*Cfg))
	}))
}

func New(cfg *Cfg) error {
	cfg.Disabled = false
	if cfg.ServiceName == "" {
		cfg.ServiceName = runenv.Project
	}

	if cfg.Sampler != nil {
		cfg.Sampler = &config.SamplerConfig{
			Type:  jaeger.SamplerTypeProbabilistic,
			Param: 0.01,
		}
	}

	trace, _, err := cfg.NewTracer(
		config.Reporter(reporter.NewIoReporter(cfg.Logger, cfg.BatchSize)),
		config.Logger(newLog("tracing")),
		config.Metrics(prometheus.New()),
	)
	xerror.Exit(err)

	opentracing.SetGlobalTracer(trace)
	return nil
}

// spanFromPHPRequest 解析php-molten组件链路
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
