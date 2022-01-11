package jaeger

import (
	"github.com/opentracing/opentracing-go"
	"github.com/pubgo/xerror"
	"github.com/uber/jaeger-client-go"
	"github.com/uber/jaeger-client-go/config"
	"github.com/uber/jaeger-client-go/rpcmetrics"
	"github.com/uber/jaeger-lib/metrics"
	jprom "github.com/uber/jaeger-lib/metrics/prometheus"

	"github.com/pubgo/lava/plugins/tracing"
	"github.com/pubgo/lava/plugins/tracing/jaeger/reporter"
	"github.com/pubgo/lava/runenv"
	"github.com/pubgo/lava/types"
	"github.com/pubgo/lava/version"
)

var _ = jaeger.NewNullReporter()

func init() {
	xerror.Exit(tracing.RegisterFactory(Name, func(cfgMap types.CfgMap) error {
		var cfg = DefaultCfg()
		cfg.ServiceName = runenv.Project
		cfg.Tags = append(cfg.Tags, opentracing.Tag{Key: "version", Value: version.Version})
		xerror.Panic(cfgMap.Decode(cfg))
		return New(cfg)
	}))
}

func New(cfg *Cfg) (err error) {
	defer xerror.RespErr(&err)

	cfg.Disabled = false
	if cfg.ServiceName == "" {
		cfg.ServiceName = runenv.Project
	}

	if cfg.Sampler != nil {
		cfg.Sampler = &config.SamplerConfig{
			Type:  jaeger.SamplerTypeProbabilistic,
			Param: 1,
		}
	}

	metricsFactory := jprom.New().
		Namespace(metrics.NSOptions{Name: runenv.Domain, Tags: nil}).
		Namespace(metrics.NSOptions{Name: runenv.Project, Tags: nil})

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
