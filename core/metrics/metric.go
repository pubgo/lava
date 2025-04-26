package metrics

import (
	"context"

	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/generic"
	"github.com/pubgo/funk/log"
	"github.com/pubgo/funk/merge"
	"github.com/pubgo/funk/version"
	lo "github.com/samber/lo"
	"github.com/uber-go/tally/v4"

	"github.com/pubgo/lava/core/lifecycle"
)

func New(m lifecycle.Lifecycle, cfg *Config, log log.Logger) Metric {
	cfg = merge.Struct(generic.Ptr(DefaultCfg()), cfg).Unwrap()

	log = log.WithName(Name)

	factory := Get(cfg.Driver)
	assert.If(factory == nil, "driver factory[%s] not found", cfg.Driver)
	opts := factory(cfg, log.WithName("driver"))
	if opts == nil {
		return tally.NoopScope
	}

	opts.Tags = Tags{"project": version.Project()}

	scope, closer := tally.NewRootScope(lo.FromPtr(opts), cfg.Interval)
	m.BeforeStop(func(ctx context.Context) error { return closer.Close() })

	registerVars(scope)
	return scope
}
