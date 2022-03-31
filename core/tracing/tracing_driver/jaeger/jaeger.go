package jaeger

import (
	"github.com/opentracing/opentracing-go"
	"github.com/pubgo/lava/core/tracing"
	"github.com/pubgo/lava/core/tracing/jaeger/reporter"
	"github.com/pubgo/xerror"
	"github.com/uber/jaeger-client-go"
	"github.com/uber/jaeger-client-go/config"
	"github.com/uber/jaeger-client-go/rpcmetrics"
	"github.com/uber/jaeger-lib/metrics"
	jprom "github.com/uber/jaeger-lib/metrics/prometheus"

	"github.com/pubgo/lava/runtime"
)

// GetSpanID 从SpanContext中获取tracerID和spanID
func GetSpanID(ctx opentracing.SpanContext) (string, string) {
	c, ok := ctx.(jaeger.SpanContext)
	if !ok {
		return "", ""
	}
	return c.TraceID().String(), c.SpanID().String()
}

var _ = jaeger.NewNullReporter()

func New(cfg Cfg) (err error) {
	defer xerror.RespErr(&err)

	cfg.Disabled = false
	if cfg.ServiceName == "" {
		cfg.ServiceName = runtime.Project
	}

	if cfg.Sampler != nil {
		cfg.Sampler = &config.SamplerConfig{
			Type:  jaeger.SamplerTypeProbabilistic,
			Param: 1,
		}
	}

	metricsFactory := jprom.New().
		Namespace(metrics.NSOptions{Name: runtime.Domain, Tags: nil}).
		Namespace(metrics.NSOptions{Name: runtime.Project, Tags: nil})

	trace, _, err := cfg.NewTracer(
		config.Reporter(reporter.NewIoReporter(cfg.Logger, cfg.BatchSize)),
		config.Logger(newLog(tracing.Name)),
		config.Metrics(metricsFactory),
		config.Observer(rpcmetrics.NewObserver(metricsFactory, rpcmetrics.DefaultNameNormalizer)),
	)
	xerror.Panic(err, "cannot initialize Jaeger Tracer")

	opentracing.SetGlobalTracer(trace)
	return nil
}
