package scheduler

import (
	"github.com/pubgo/dix"
	"github.com/pubgo/xerror"

	"github.com/pubgo/lava/plugin"
)

const Name = "scheduler"

func init() {
	plugin.Register(&plugin.Base{
		Name: Name,
		OnInit: func(ent plugin.Entry) {
			quart.scheduler.Start()
			xerror.Panic(dix.Provider(quart))
			ent.AfterStop(quart.scheduler.Stop)
		},
	})
}
