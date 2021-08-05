package prometheus

import (
	"github.com/pubgo/lug/logutil"
	"github.com/pubgo/lug/metric"

	"github.com/pubgo/x/merge"
	"github.com/pubgo/xerror"
	"github.com/pubgo/xlog"
	"github.com/uber-go/tally"
	"github.com/uber-go/tally/prometheus"
)

const Name = "prometheus"

func init() {
	xerror.Exit(metric.Register(Name, func(cfg map[string]interface{}, opts *tally.ScopeOptions) (err error) {
		opts.Separator = prometheus.DefaultSeparator
		opts.SanitizeOptions = &prometheus.DefaultSanitizerOpts

		var proCfg = prometheus.Configuration{}
		xerror.Panic(merge.MapStruct(&cfg, &proCfg))

		opts.CachedReporter, err = proCfg.NewReporter(
			prometheus.ConfigurationOptions{
				OnError: func(e error) {
					xlog.Error(
						"metric error",
						logutil.Err(e),
						logutil.Pkg("metric.prometheus"),
					)
				},
			})
		return xerror.Wrap(err)
	}))
}
