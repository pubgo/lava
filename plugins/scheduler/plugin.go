package scheduler

import (
	"github.com/pubgo/dix"
	"github.com/pubgo/xerror"

	"github.com/pubgo/lava/inject"
	"github.com/pubgo/lava/plugin"
)

const Name = "scheduler"

func init() {
	plugin.Register(&plugin.Base{
		Name: Name,
		OnInit: func(p plugin.Process) {
			quart.scheduler.Start()
			xerror.Panic(dix.Provider(quart))
			inject.Register(quart, inject.WithVal(quart))
			p.BeforeStop(quart.scheduler.Stop)
		},
	})
}
