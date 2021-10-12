package automaxprocs

import (
	"github.com/pubgo/xerror"
	"go.uber.org/automaxprocs/maxprocs"
	"go.uber.org/zap"

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
				zap.L().WithOptions(zap.AddCallerSkip(2)).Named(name).Sugar().Infof(s, i...)
			}
			xerror.ExitErr(maxprocs.Set(maxprocs.Logger(handler))).(func())()
		},
	})
}
