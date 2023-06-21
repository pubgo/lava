package metrics

import (
	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/generic"
	"github.com/pubgo/funk/log"
	"github.com/pubgo/funk/merge"
	"github.com/pubgo/funk/version"
	"github.com/uber-go/tally/v4"

	"github.com/pubgo/lava/core/lifecycle"
)

func New(m lifecycle.Lifecycle, cfg *Config, log log.Logger) Metric {
	cfg = merge.Struct(generic.Ptr(DefaultCfg()), cfg).Unwrap()

	log = log.WithName(Name)

	factory := Get(cfg.Driver)
	assert.If(factory == nil, "driver factory[%s] not found", cfg.Driver)
	opts := factory(cfg, log)
	if opts == nil {
		return tally.NoopScope
	}

	opts.Tags = Tags{"project": version.Project()}

	scope, closer := tally.NewRootScope(*opts, cfg.Interval)
	m.BeforeStop(func() { assert.Must(closer.Close()) })

	registerVars(scope)
	return scope
}
