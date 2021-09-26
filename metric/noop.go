package metric

import (
	"github.com/uber-go/tally"
)

func init() {
	Register("noop", func(cfg map[string]interface{}, opts *tally.ScopeOptions) error {
		opts.Reporter = tally.NullStatsReporter
		return nil
	})
}
