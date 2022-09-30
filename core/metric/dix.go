package metric

import (
	"context"

	"github.com/pubgo/dix/di"
	"github.com/pubgo/funk/assert"
	"github.com/uber-go/tally/v4"

	"github.com/pubgo/lava/config"
	"github.com/pubgo/lava/core/lifecycle"
	"github.com/pubgo/lava/core/runmode"
	"github.com/pubgo/lava/logging/logkey"
	"github.com/pubgo/lava/service"
)

func init() {
	di.Provide(func(c config.Config) *Cfg {
		var cfg = DefaultCfg()
		assert.Must(c.UnmarshalKey(Name, &cfg))
		assert.If(cfg.Driver == "", "metric driver is null")
		return &cfg
	})

	di.Provide(func(m lifecycle.Lifecycle, cfg *Cfg, optMap map[string]*tally.ScopeOptions) Metric {
		var opts = optMap[cfg.Driver]
		if opts == nil {
			opts = &tally.ScopeOptions{Reporter: tally.NullStatsReporter}
		}
		opts.Tags = Tags{logkey.Project: runmode.Project}
		if cfg.Separator != "" {
			opts.Separator = cfg.Separator
		}

		scope, closer := tally.NewRootScope(*opts, cfg.Interval)
		m.BeforeStop(func() { assert.Must(closer.Close()) })

		registerVars(scope)
		return scope
	})

	di.Provide(func(m Metric) service.Middleware {
		return func(next service.HandlerFunc) service.HandlerFunc {
			return func(ctx context.Context, req service.Request, resp service.Response) error {
				return next(CreateCtxWithMetric(ctx, m), req, resp)
			}
		}
	})
}
