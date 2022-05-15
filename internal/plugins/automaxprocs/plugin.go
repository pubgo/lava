package automaxprocs

import (
	"fmt"

	"github.com/pubgo/xerror"
	"go.uber.org/automaxprocs/maxprocs"
	"go.uber.org/fx"

	"github.com/pubgo/lava/core/lifecycle"
	"github.com/pubgo/lava/logging"
)

func Module() fx.Option {
	return fx.Invoke(func(m lifecycle.Lifecycle) {
		m.BeforeStops(func() {
			const name = "automaxprocs"
			var log = func(s string, i ...interface{}) { logging.Component(name).Depth(2).Info(fmt.Sprintf(s, i...)) }
			xerror.ExitErr(maxprocs.Set(maxprocs.Logger(log))).(func())()
		})
	})
}
