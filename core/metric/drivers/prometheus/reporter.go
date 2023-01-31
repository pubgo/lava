package prometheus

import (
	"github.com/pubgo/funk/assert"
	tally "github.com/uber-go/tally/v4"
	"github.com/uber-go/tally/v4/prometheus"
	"go.uber.org/zap"

	"github.com/pubgo/lava/core/metric"
	"github.com/pubgo/lava/logging"
	"github.com/pubgo/lava/logging/logkey"
	"github.com/pubgo/lava/logging/logutil"
)

const Name = "prometheus"
const urlPath = "/metrics"

func New(conf *metric.Cfg, log *logging.Logger) map[string]*tally.ScopeOptions {
	var logs = logging.ModuleLog(log, logutil.Names(metric.Name, Name))

	if conf.Driver != Name {
		return nil
	}

	opts := tally.ScopeOptions{}
	opts.Separator = prometheus.DefaultSeparator
	opts.SanitizeOptions = &prometheus.DefaultSanitizerOpts

	var proCfg = &prometheus.Configuration{}
	if conf.DriverCfg != nil {
		assert.Must(conf.DriverCfg.Decode(proCfg))
	}

	reporter := assert.Must1(proCfg.NewReporter(
		prometheus.ConfigurationOptions{
			OnError: func(e error) {
				logs.WithErr(e, zap.Any(logkey.Config, conf)).Error("metric.prometheus init error")
			},
		},
	))
	debug.Get(urlPath, debug.Wrap(reporter.HTTPHandler()))

	opts.CachedReporter = reporter
	return map[string]*tally.ScopeOptions{Name: &opts}
}
