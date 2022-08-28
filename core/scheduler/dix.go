package scheduler

import (
	"github.com/pubgo/dix/di"
	"github.com/pubgo/lava/core/lifecycle"
	"github.com/pubgo/lava/logging"
)

const Name = "scheduler"

func init() {
	di.Provide(func(m lifecycle.Lifecycle, log *logging.Logger) *Scheduler {
		var quart = New(log)
		quart.Start()
		m.BeforeStop(quart.Stop)
		return quart
	})
}
