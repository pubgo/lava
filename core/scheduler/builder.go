package scheduler

import (
	"github.com/pubgo/lava/core/lifecycle"
	"github.com/pubgo/lava/inject"
)

const Name = "scheduler"

func init() {
	inject.Provide(func(m lifecycle.Lifecycle) *Scheduler {
		quart.scheduler.Start()
		m.BeforeStops(quart.scheduler.Stop)
		return quart
	})
}
