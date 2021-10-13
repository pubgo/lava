package scheduler

import (
	"github.com/pubgo/lava/plugin"
)

const Name = "scheduler"

func init() {
	plugin.Register(&plugin.Base{
		Name: Name,
		OnInit: func(ent plugin.Entry) {
			ent.BeforeStart(func() {
				quart.scheduler.Start()
			})
			ent.AfterStop(func() {
				quart.scheduler.Stop()
			})
		},
	})
}
