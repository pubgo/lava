package prometheus

import (
	"github.com/pubgo/xerror"
	"github.com/uber-go/tally"
	"github.com/uber-go/tally/prometheus"
	"go.uber.org/zap"

	"github.com/pubgo/lava/logger"
	"github.com/pubgo/lava/logger/logkey"
	"github.com/pubgo/lava/logger/logutil"
	"github.com/pubgo/lava/mux"
	"github.com/pubgo/lava/plugins/metric"
	"github.com/pubgo/lava/types"
)

const Name = "prometheus"
const urlPath = "/metrics"

var logs = logger.Component(logutil.Names(metric.Name, Name))

func init() {
	metric.RegisterFactory(Name, func(cfg types.CfgMap, opts *tally.ScopeOptions) (err error) {
		defer xerror.RespErr(&err)

		opts.Separator = prometheus.DefaultSeparator
		opts.SanitizeOptions = &prometheus.DefaultSanitizerOpts

		var proCfg = &prometheus.Configuration{}
		xerror.Panic(cfg.Decode(proCfg))
		reporter, err1 := proCfg.NewReporter(
			prometheus.ConfigurationOptions{
				OnError: func(e error) {
					logs.WithErr(e, zap.Any(logkey.Config, cfg)).Error("metric.prometheus init error")
				},
			},
		)
		xerror.Panic(err1)
		mux.DebugGet(urlPath, reporter.HTTPHandler().ServeHTTP)

		opts.CachedReporter = reporter
		return nil
	})
}
