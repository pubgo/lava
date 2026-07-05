package tracingbuilder

import (
	"github.com/pubgo/funk/v2/merge"
	lo "github.com/samber/lo"

	"github.com/pubgo/lava/v2/core/lifecycle"
	"github.com/pubgo/lava/v2/core/tracing"
)

// New wires OpenTelemetry via DI, merging user config with tracing defaults.
func New(lc lifecycle.Lifecycle, cfg *tracing.Config) tracing.Provider {
	merged := merge.Struct(lo.ToPtr(tracing.DefaultCfg()), cfg).Unwrap()
	return tracing.NewProvider(merged, lc)
}
