package automaxprocs

import (
	"github.com/pubgo/xerror"
	"go.uber.org/automaxprocs/maxprocs"

	"github.com/pubgo/lava/internal/logz"
	"github.com/pubgo/lava/plugin"
)

func init() {
	const name = "automaxprocs"
	plugin.Register(&plugin.Base{
		Name:       name,
		Url:        "https://pkg.go.dev/go.uber.org/automaxprocs",
		Descriptor: "Automatically set GOMAXPROCS to match Linux container CPU quota.",
		OnInit: func(ent plugin.Entry) {
			var handler = func(s string, i ...interface{}) {
				logz.Named(name, 2).Infof(s, i...)
			}
			xerror.ExitErr(maxprocs.Set(maxprocs.Logger(handler))).(func())()
		},
	})
}
