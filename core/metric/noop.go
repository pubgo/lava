package metric

import (
	"github.com/pubgo/lava/config/config_type"
	"github.com/uber-go/tally"
)

func init() {
	RegisterFactory("noop", func(cfg config_type.CfgMap, opts *tally.ScopeOptions) error {
		opts.Reporter = tally.NullStatsReporter
		return nil
	})
}
