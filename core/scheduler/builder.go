package scheduler

import (
	"github.com/pubgo/dix"

	"github.com/pubgo/lava/core/lifecycle"
	"github.com/pubgo/lava/logging"
)

const Name = "scheduler"

func init() {
	dix.Provider(func(m lifecycle.Lifecycle, log *logging.Logger) *Scheduler {
		quart.log = log.Named(Name)
		quart.scheduler.Start()
		m.BeforeStops(quart.scheduler.Stop)
		return quart
	})
}
