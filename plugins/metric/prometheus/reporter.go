package prometheus

import (
	"github.com/pubgo/xerror"
	"github.com/uber-go/tally"
	"github.com/uber-go/tally/prometheus"

	"github.com/pubgo/lava/internal/logz"
	"github.com/pubgo/lava/mux"
	"github.com/pubgo/lava/pkg/merge"
	"github.com/pubgo/lava/plugins/logger"
	"github.com/pubgo/lava/plugins/metric"
)

const Name = "prometheus"

func init() {
	metric.Register(Name, func(cfg map[string]interface{}, opts *tally.ScopeOptions) error {
		opts.Separator = prometheus.DefaultSeparator
		opts.SanitizeOptions = &prometheus.DefaultSanitizerOpts

		proCfg := merge.MapStruct(&prometheus.Configuration{}, &cfg).(*prometheus.Configuration)
		reporter, err := proCfg.NewReporter(
			prometheus.ConfigurationOptions{
				OnError: func(e error) {
					logz.With(Name, logger.WithErr(e)...).Errorf("metric.prometheus error")
				},
			},
		)
		xerror.Panic(err)
		mux.Handle("/metrics", reporter.HTTPHandler())

		opts.CachedReporter = reporter
		return nil
	})
}
