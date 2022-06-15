package automaxprocs

import (
	"fmt"

	"github.com/pubgo/dix"
	"github.com/pubgo/xerror"
	"go.uber.org/automaxprocs/maxprocs"

	"github.com/pubgo/lava/core/lifecycle"
	"github.com/pubgo/lava/logging"
)

func init() {
	dix.Provider(func() lifecycle.Handler {
		return func(lc lifecycle.Lifecycle) {
			const name = "automaxprocs"
			var log = func(s string, i ...interface{}) { logging.GetGlobal(name).Depth(2).Info(fmt.Sprintf(s, i...)) }
			undo := xerror.ExitErr(maxprocs.Set(maxprocs.Logger(log))).(func())
			lc.BeforeStops(undo)
		}
	})
}
