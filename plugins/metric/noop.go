package metric

import (
	"github.com/pubgo/lava/types"
	"github.com/uber-go/tally"
)

func init() {
	RegisterFactory("noop", func(cfg types.CfgMap, opts *tally.ScopeOptions) error {
		opts.Reporter = tally.NullStatsReporter
		return nil
	})
}
