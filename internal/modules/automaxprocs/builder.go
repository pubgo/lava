package automaxprocs

import (
	"fmt"

	"github.com/pubgo/dix"
	"github.com/pubgo/funk/assert"
	"go.uber.org/automaxprocs/maxprocs"

	"github.com/pubgo/lava/core/lifecycle"
	"github.com/pubgo/lava/logging"
)

const name = "automaxprocs"

func init() {
	dix.Provider(func() lifecycle.Handler {
		return func(lc lifecycle.Lifecycle) {
			var log = func(s string, i ...interface{}) { logging.GetGlobal(name).Depth(2).Info(fmt.Sprintf(s, i...)) }
			assert.Must1(maxprocs.Set(maxprocs.Logger(log)))
		}
	})
}
