package automaxprocs

import (
	"github.com/pubgo/xerror"
	"go.uber.org/automaxprocs/maxprocs"
	"go.uber.org/zap"

	"github.com/pubgo/lug/plugin"
)

func init() {
	plugin.Register(&plugin.Base{
		Name: "automaxprocs",
		OnInit: func(ent plugin.Entry) {
			logs := zap.L().WithOptions(zap.AddCallerSkip(2))
			var log = maxprocs.Logger(func(s string, i ...interface{}) { logs.Sugar().Infof(s, i...) })
			xerror.ExitErr(maxprocs.Set(log)).(func())()
		},
	})
}
