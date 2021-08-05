package metric

import (
	"github.com/pubgo/xerror"
	"github.com/uber-go/tally"
)

func init() {
	xerror.Exit(Register("noop", func(cfg map[string]interface{}, opts *tally.ScopeOptions) error {
		opts.Reporter = tally.NullStatsReporter
		return nil
	}))
}
