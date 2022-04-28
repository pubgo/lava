package scheduler

import (
	"github.com/pubgo/lava/module"
	"github.com/pubgo/lava/running"
)

const Name = "scheduler"

func init() {
	module.Provide(func(run running.Running) *Scheduler {
		run.AfterStarts(quart.scheduler.Start)
		run.BeforeStops(quart.scheduler.Stop)
		return quart
	})
}
