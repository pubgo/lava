package metricbuilder

import (
	"context"

	"github.com/pubgo/funk/v2/assert"
	"github.com/pubgo/funk/v2/buildinfo/version"
	"github.com/pubgo/funk/v2/log"
	"github.com/pubgo/funk/v2/merge"
	lo "github.com/samber/lo"
	"github.com/uber-go/tally/v4"

	"github.com/pubgo/lava/v2/core/lifecycle"
	"github.com/pubgo/lava/v2/core/metrics"
)

func New(m lifecycle.Lifecycle, cfg *metrics.Config, log log.Logger) metrics.Metric {
	cfg = merge.Struct(lo.ToPtr(metrics.DefaultCfg()), cfg).Unwrap()

	log = log.WithName(metrics.Name)

	factory := metrics.Get(cfg.Driver)
	assert.If(factory == nil, "driver factory[%s] not found", cfg.Driver)
	opts := factory(cfg, log.WithName("driver"))
	if opts == nil {
		return tally.NoopScope
	}

	opts.Tags = metrics.Tags{"project": version.Project()}

	scope, closer := tally.NewRootScope(lo.FromPtr(opts), cfg.Interval)
	m.BeforeStop(func(ctx context.Context) error { return closer.Close() })

	registerVars(scope)
	return scope
}
