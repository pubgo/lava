package jaeger

import (
	"github.com/pubgo/lug/tracing"
	"github.com/pubgo/x/merge"
	"github.com/pubgo/xerror"
	"github.com/pubgo/xlog"
	"github.com/uber/jaeger-client-go"
	"github.com/uber/jaeger-client-go/config"
	"github.com/uber/jaeger-lib/metrics/prometheus"
	"go.uber.org/zap"
)

func init() {
	xerror.Exit(tracing.Register(Name, func(cfgMap map[string]interface{}) (tracing.Tracer, error) {
		var cfg = GetDefaultCfg()
		xerror.Panic(merge.MapStruct(&cfg, cfgMap))
		return New(cfg)
	}))
}

func New(cfg *Cfg) (tracing.Tracer, error) {
	var logs = &logger{logs: xlog.GetLogger(cfg.ServiceName,
		zap.AddCallerSkip(4),
		zap.Fields(zap.String("type", "tracing")))}

	logOpt := config.Logger(logs)
	metricOpt := config.Metrics(prometheus.New())
	reporter := config.Reporter(logs)
	sampler := config.Sampler(jaeger.NewConstSampler(true))

	var tracer tracing.Tracer
	trace, closer, err := cfg.NewTracer(
		logOpt,
		metricOpt,
		reporter,
		sampler,
	)
	xerror.Panic(err)
	tracer.Tracer = trace
	tracer.Closer = closer

	return tracer, nil
}
