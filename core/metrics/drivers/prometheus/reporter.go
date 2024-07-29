package prometheus

import (
	"github.com/prometheus/common/model"
	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/log"
	tally "github.com/uber-go/tally/v4"
	"github.com/uber-go/tally/v4/prometheus"

	"github.com/pubgo/lava/core/debug"
	"github.com/pubgo/lava/core/metrics"
)

const (
	Name    = "prometheus"
	urlPath = "/metrics"
)

func init() {
	metrics.Register(Name, New)
}

func New(conf *metrics.Config, log log.Logger) *tally.ScopeOptions {
	if conf.Driver != Name {
		return nil
	}

	opts := tally.ScopeOptions{}
	opts.Separator = prometheus.DefaultSeparator
	//opts.SanitizeOptions = &prometheus.DefaultSanitizerOpts
	model.NameValidationScheme = model.UTF8Validation

	proCfg := &prometheus.Configuration{TimerType: "histogram"}

	if conf.DriverCfg != nil {
		assert.Must(conf.DriverCfg.Decode(proCfg))
	}

	logs := log.WithName(Name)
	reporter := assert.Must1(proCfg.NewReporter(
		prometheus.ConfigurationOptions{
			OnError: func(err error) {
				logs.Err(err).Any("metric-config", conf).Msg("metric.prometheus init error")
			},
		},
	))
	debug.Get(urlPath, debug.Wrap(reporter.HTTPHandler()))

	opts.CachedReporter = reporter
	return &opts
}
