package jaeger

import (
	"github.com/opentracing/opentracing-go"
	"github.com/pubgo/xerror"
	"github.com/uber/jaeger-client-go"
	"github.com/uber/jaeger-client-go/config"
	"github.com/uber/jaeger-client-go/rpcmetrics"
	"github.com/uber/jaeger-lib/metrics"
	jprom "github.com/uber/jaeger-lib/metrics/prometheus"

	"github.com/pubgo/lava/logger/logkey"
	"github.com/pubgo/lava/plugins/tracing"
	"github.com/pubgo/lava/plugins/tracing/jaeger/reporter"
	"github.com/pubgo/lava/runtime"
	"github.com/pubgo/lava/types"
	"github.com/pubgo/lava/version"
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

func init() {
	xerror.Exit(tracing.RegisterFactory(Name, func(cfgMap types.CfgMap) error {
		defer func() {
			tracing.GetSpanID = GetSpanID
		}()

		var cfg = DefaultCfg()
		cfg.ServiceName = runtime.Project
		cfg.Tags = append(cfg.Tags, opentracing.Tag{Key: logkey.Version, Value: version.Version})
		xerror.Panic(cfgMap.Decode(cfg))
		return New(cfg)
	}))
}

func New(cfg *Cfg) (err error) {
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
