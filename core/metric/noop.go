package metric

import (
	"github.com/pubgo/lava/config"
	"github.com/uber-go/tally"
)

func init() {
	RegisterFactory("noop", func(cfg config.CfgMap, opts *tally.ScopeOptions) error {
		opts.Reporter = tally.NullStatsReporter
		return nil
	})
}
