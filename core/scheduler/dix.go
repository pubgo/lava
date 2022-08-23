package scheduler

import (
	"github.com/pubgo/dix"
	"github.com/pubgo/lava/core/lifecycle"
	"github.com/pubgo/lava/logging"
)

const Name = "scheduler"

func init() {
	dix.Provider(func(m lifecycle.Lifecycle, log *logging.Logger) *Scheduler {
		var quart = New(log)
		quart.Start()
		m.BeforeStop(quart.Stop, "scheduler close")
		return quart
	})
}
