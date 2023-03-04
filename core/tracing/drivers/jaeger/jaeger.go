package jaeger

import (
	"github.com/opentracing/opentracing-go"
	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/recovery"
	"github.com/pubgo/funk/runmode"
	"github.com/uber/jaeger-client-go"
	"github.com/uber/jaeger-client-go/config"
	"github.com/uber/jaeger-client-go/rpcmetrics"
	"github.com/uber/jaeger-lib/metrics"
	jprom "github.com/uber/jaeger-lib/metrics/prometheus"

	"github.com/pubgo/lava/core/tracing"
	"github.com/pubgo/lava/core/tracing/drivers/jaeger/reporter"
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
	defer recovery.Err(&err)

	cfg.Disabled = false
	if cfg.ServiceName == "" {
		cfg.ServiceName = runmode.Project
	}

	if cfg.Sampler != nil {
		cfg.Sampler = &config.SamplerConfig{
			Type:  jaeger.SamplerTypeProbabilistic,
			Param: 1,
		}
	}

	metricsFactory := jprom.New().
		Namespace(metrics.NSOptions{Name: runmode.Project, Tags: nil})

	trace, _, err := cfg.NewTracer(
		config.Reporter(reporter.NewIoReporter(cfg.Logger, cfg.BatchSize)),
		config.Logger(newLog(tracing.Name)),
		config.Metrics(metricsFactory),
		config.Observer(rpcmetrics.NewObserver(metricsFactory, rpcmetrics.DefaultNameNormalizer)),
	)
	assert.Must(err, "cannot initialize Jaeger Tracer")

	opentracing.SetGlobalTracer(trace)
	return nil
}
