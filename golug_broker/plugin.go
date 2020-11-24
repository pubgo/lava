package golug_broker

import (
	"github.com/asim/nitro/v3/config/reader"
	"github.com/pubgo/catdog/catdog_plugin"
	"github.com/pubgo/dix/dix_run"
	"github.com/pubgo/xerror"
	"github.com/spf13/pflag"
)

func init() {
	xerror.Exit(catdog_plugin.Register(&catdog_plugin.Base{
		Name: "broker",
		OnFlags: func(flags *pflag.FlagSet) {
		},
		OnInit: func(r reader.Value) {
			xerror.Exit(dix_run.WithBeforeStart(func(ctx *dix_run.BeforeStartCtx) {
			}))
		},
	}))
}

//xerror.Exit(r.s.Init(server.Broker(catdog_broker.Default.Broker)))
