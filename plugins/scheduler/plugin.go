package scheduler

import (
	"github.com/pubgo/lava/plugin"
)

const Name = "scheduler"

func init() {
	plugin.Register(&plugin.Base{
		Name: Name,
		OnInit: func(p plugin.Process) {
			quart.scheduler.Start()
			p.BeforeStop(func() { quart.scheduler.Stop() })
		},
	})
}
