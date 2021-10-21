package scheduler

import (
	"github.com/pubgo/lava/plugin"
	"github.com/pubgo/lava/resource"
)

const Name = "scheduler"

func init() {
	plugin.Register(&plugin.Base{
		Name: Name,
		OnInit: func(ent plugin.Entry) {
			quart.scheduler.Start()
			resource.Update("", quart)
			ent.BeforeStop(quart.scheduler.Stop)
		},
	})
}
