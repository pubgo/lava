package prometheus

import (
	"github.com/pubgo/dix"
	"github.com/pubgo/xerror"
	"github.com/uber-go/tally"
	"github.com/uber-go/tally/prometheus"
	"go.uber.org/zap"

	"github.com/pubgo/lava/core/metric"
	"github.com/pubgo/lava/debug"
	"github.com/pubgo/lava/logging"
	"github.com/pubgo/lava/logging/logkey"
	"github.com/pubgo/lava/logging/logutil"
)

const Name = "prometheus"
const urlPath = "/metrics"

func init() {
	dix.Provider(func(conf *metric.Cfg, log *logging.Logger) map[string]*tally.ScopeOptions {
		var logs = logging.ModuleLog(log, logutil.Names(metric.Name, Name))

		if conf.Driver != Name {
			return nil
		}

		opts := tally.ScopeOptions{}
		opts.Separator = prometheus.DefaultSeparator
		opts.SanitizeOptions = &prometheus.DefaultSanitizerOpts

		var proCfg = &prometheus.Configuration{}
		if conf.DriverCfg != nil {
			xerror.Panic(conf.DriverCfg.Decode(proCfg))
		}

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
		return map[string]*tally.ScopeOptions{Name: &opts}
	})
}
