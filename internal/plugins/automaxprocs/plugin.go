package automaxprocs

import (
	"github.com/pubgo/xerror"
	"go.uber.org/automaxprocs/maxprocs"

	"github.com/pubgo/lava/logz"
	"github.com/pubgo/lava/plugin"
)

func init() {
	const name = "automaxprocs"
	var logs = logz.Component(name)
	plugin.Register(&plugin.Base{
		Name:       name,
		Url:        "https://pkg.go.dev/go.uber.org/automaxprocs",
		Descriptor: "Automatically set GOMAXPROCS to match Linux container CPU quota.",
		OnInit: func(p plugin.Process) {
			var l = maxprocs.Logger(func(s string, i ...interface{}) { logs.DepthS(2).Infof(s, i...) })
			xerror.ExitErr(maxprocs.Set(l)).(func())()
		},
	})
}
