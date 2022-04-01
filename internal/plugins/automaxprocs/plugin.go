package automaxprocs

import (
	"fmt"

	"github.com/pubgo/xerror"
	"go.uber.org/automaxprocs/maxprocs"

	"github.com/pubgo/lava/core/logging"
	"github.com/pubgo/lava/plugin"
)

func init() {
	const name = "automaxprocs"
	var log = func(s string, i ...interface{}) { logging.Component(name).Depth(2).Info(fmt.Sprintf(s, i...)) }

	plugin.Register(&plugin.Base{
		Name:  name,
		Url:   "https://pkg.go.dev/go.uber.org/automaxprocs",
		Short: "Automatically set GOMAXPROCS to match Linux container CPU quota.",
		OnInit: func(p plugin.Process) {
			xerror.ExitErr(maxprocs.Set(maxprocs.Logger(log))).(func())()
		},
	})
}
