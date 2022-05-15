package scheduler

import (
	"go.uber.org/fx"

	"github.com/pubgo/lava/core/lifecycle"
	"github.com/pubgo/lava/inject"
)

const Name = "scheduler"

func init() {
	inject.Register(fx.Provide(func(m lifecycle.Lifecycle) *Scheduler {
		quart.scheduler.Start()
		m.BeforeStops(quart.scheduler.Stop)
		return quart
	}))
}
