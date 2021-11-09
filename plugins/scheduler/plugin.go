package scheduler

import (
	"github.com/pubgo/lava/plugin"
	"github.com/pubgo/lava/resource"
)

const Name = "scheduler"

func init() {
	plugin.Register(&plugin.Base{
		Name: Name,
		OnInit: func(p plugin.Process) {
			quart.scheduler.Start()
			resource.Update("", quart)
			p.BeforeStop(quart.scheduler.Stop)
		},
	})
}
