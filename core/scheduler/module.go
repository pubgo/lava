package scheduler

import (
	"go.uber.org/fx"

	"github.com/pubgo/lava/module"
	"github.com/pubgo/lava/service"
)

const Name = "scheduler"

func init() {
	module.Register(fx.Provide(func(srv service.Service) *Scheduler {
		quart.scheduler.Start()
		srv.BeforeStops(quart.scheduler.Stop)
		return quart
	}))
}
