package automaxprocs

import (
	"github.com/pubgo/lug/entry"
	"github.com/pubgo/lug/plugin"

	"github.com/pubgo/xerror"
	"go.uber.org/automaxprocs/maxprocs"
	"go.uber.org/zap"
)

func init() {
	plugin.Register(&plugin.Base{
		Name: "automaxprocs",
		OnInit: func(ent entry.Entry) {
			logs := zap.L().WithOptions(zap.AddCallerSkip(2))
			var log = maxprocs.Logger(func(s string, i ...interface{}) { logs.Sugar().Infof(s, i...) })
			xerror.ExitErr(maxprocs.Set(log)).(func())()
		},
	})
}
