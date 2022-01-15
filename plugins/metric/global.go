package metric

import (
	"sync/atomic"

	"github.com/uber-go/tally"
)

var g atomic.Value

func GetGlobal() tally.Scope {
	var val = g.Load()
	if val == nil {
		return tally.NoopScope
	}

	return val.(tally.Scope)
}
