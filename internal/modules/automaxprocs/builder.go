package automaxprocs

import (
	"fmt"

	"github.com/pubgo/dix"
	"github.com/pubgo/lava/core/lifecycle"
	"github.com/pubgo/lava/logging"
	"github.com/pubgo/xerror"
	"go.uber.org/automaxprocs/maxprocs"
)

func init() {
	dix.Register(func(m lifecycle.Lifecycle) {
		const name = "automaxprocs"
		var log = func(s string, i ...interface{}) { logging.Component(name).Depth(2).Info(fmt.Sprintf(s, i...)) }
		undo := xerror.ExitErr(maxprocs.Set(maxprocs.Logger(log))).(func())
		m.BeforeStops(undo)
	})
}
