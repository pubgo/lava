package scheduler

import (
	"go.uber.org/fx"

	"github.com/pubgo/lava/module"
	"github.com/pubgo/lava/running"
)

const Name = "scheduler"

func init() {
	module.Register(fx.Provide(func(run running.Running) *Scheduler {
		run.AfterStarts(quart.scheduler.Start)
		run.BeforeStops(quart.scheduler.Stop)
		return quart
	}))
}
