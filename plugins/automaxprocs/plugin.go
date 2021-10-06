package automaxprocs

import (
	"github.com/pubgo/xerror"
	"go.uber.org/automaxprocs/maxprocs"
	"go.uber.org/zap"

	"github.com/pubgo/lug/logger"
	"github.com/pubgo/lug/plugin"
)

func init() {
	const name = "automaxprocs"
	plugin.Register(&plugin.Base{
		Name:       name,
		Url:        "https://pkg.go.dev/go.uber.org/automaxprocs",
		Descriptor: "Automatically set GOMAXPROCS to match Linux container CPU quota.",
		OnInit: func(ent plugin.Entry) {
			var logs = zap.L()
			logger.On(func(log *zap.Logger) { logs = log.Named(name).WithOptions(zap.AddCallerSkip(2)) })
			var handler = func(s string, i ...interface{}) { logs.Sugar().Infof(s, i...) }
			xerror.ExitErr(maxprocs.Set(maxprocs.Logger(handler))).(func())()
		},
	})
}
