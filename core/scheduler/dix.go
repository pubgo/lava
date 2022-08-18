package scheduler

import (
	"github.com/pubgo/dix"
	"github.com/pubgo/funk/recovery"

	"github.com/pubgo/lava/core/lifecycle"
	"github.com/pubgo/lava/logging"
)

const Name = "scheduler"

func init() {
	defer recovery.Exit()

	dix.Provider(func(m lifecycle.Lifecycle, log *logging.Logger) *Scheduler {
		quart.log = log.Named(Name)
		quart.scheduler.Start()
		m.BeforeStop(func() error { quart.scheduler.Stop(); return nil }, "scheduler close")
		return quart
	})
}
