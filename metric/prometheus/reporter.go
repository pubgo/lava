package prometheus

import (
	"github.com/pubgo/x/merge"
	"github.com/pubgo/xerror"
	"github.com/uber-go/tally"
	"github.com/uber-go/tally/prometheus"

	"github.com/pubgo/lava/logger"
	"github.com/pubgo/lava/logz"
	"github.com/pubgo/lava/metric"
)

const Name = "prometheus"

func init() {
	metric.Register(Name, func(cfg map[string]interface{}, opts *tally.ScopeOptions) (err error) {
		opts.Separator = prometheus.DefaultSeparator
		opts.SanitizeOptions = &prometheus.DefaultSanitizerOpts

		var proCfg = prometheus.Configuration{}
		xerror.Panic(merge.MapStruct(&cfg, &proCfg))

		opts.CachedReporter, err = proCfg.NewReporter(
			prometheus.ConfigurationOptions{
				OnError: func(e error) {
					logz.With(Name, logger.WithErr(e)...).Errorf("metric.prometheus error")
				},
			})
		return xerror.Wrap(err)
	})
}
