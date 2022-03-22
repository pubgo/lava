package prometheus

import (
	"github.com/pubgo/xerror"
	"github.com/uber-go/tally"
	"github.com/uber-go/tally/prometheus"
	"go.uber.org/zap"

	"github.com/pubgo/lava/config/config_type"
	"github.com/pubgo/lava/debug"
	"github.com/pubgo/lava/logging"
	"github.com/pubgo/lava/logging/logkey"
	"github.com/pubgo/lava/logging/logutil"
	"github.com/pubgo/lava/plugins/metric"
)

const Name = "prometheus"
const urlPath = "/metrics"

var logs = logging.Component(logutil.Names(metric.Name, Name))

func init() {
	metric.RegisterFactory(Name, func(cfg config_type.CfgMap, opts *tally.ScopeOptions) (err error) {
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
		debug.Get(urlPath, debug.Wrap(reporter.HTTPHandler()))

		opts.CachedReporter = reporter
		return nil
	})
}
