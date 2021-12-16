package metric

import (
	"github.com/uber-go/tally"
)

func init() {
	RegisterFactory("noop", func(cfg map[string]interface{}, opts *tally.ScopeOptions) error {
		opts.Reporter = tally.NullStatsReporter
		return nil
	})
}
