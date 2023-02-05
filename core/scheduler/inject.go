package scheduler

import (
	"github.com/pubgo/dix/di"
	"github.com/pubgo/funk/lifecycle"
	"github.com/pubgo/funk/log"
)

const Name = "scheduler"

func init() {
	di.Provide(func(m lifecycle.Lifecycle, log log.Logger) *Scheduler {
		var quart = New(log)
		quart.Start()
		m.BeforeStop(quart.Stop)
		return quart
	})
}
