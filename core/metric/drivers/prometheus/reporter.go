package prometheus

import (
	"github.com/pubgo/xerror"
	"github.com/uber-go/tally"
	"github.com/uber-go/tally/prometheus"
	"go.uber.org/zap"

	metric "github.com/pubgo/lava/core/metric"
	"github.com/pubgo/lava/debug"
	"github.com/pubgo/lava/inject"
	"github.com/pubgo/lava/logging"
	"github.com/pubgo/lava/logging/logkey"
	"github.com/pubgo/lava/logging/logutil"
)

const Name = "prometheus"
const urlPath = "/metrics"

var logs = logging.Component(logutil.Names(metric.Name, Name))

func init() {
	inject.RegGroup(metric.Name, func(conf *metric.Cfg) *tally.ScopeOptions {
		if conf.Driver != Name || conf.DriverCfg == nil {
			return nil
		}

		opts := tally.ScopeOptions{}
		opts.Separator = prometheus.DefaultSeparator
		opts.SanitizeOptions = &prometheus.DefaultSanitizerOpts

		var proCfg = &prometheus.Configuration{}
		xerror.Panic(conf.DriverCfg.Decode(proCfg))
		reporter, err1 := proCfg.NewReporter(
			prometheus.ConfigurationOptions{
				OnError: func(e error) {
					logs.WithErr(e, zap.Any(logkey.Config, conf)).Error("metric.prometheus init error")
				},
			},
		)
		xerror.Panic(err1)
		debug.Get(urlPath, debug.Wrap(reporter.HTTPHandler()))

		opts.CachedReporter = reporter
		return &opts
	})
}
