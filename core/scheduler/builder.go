package scheduler

import (
	"github.com/pubgo/funk/log"
	"github.com/reugn/go-quartz/quartz"

	"github.com/pubgo/lava/core/lifecycle"
)

const Name = "scheduler"

func New(m lifecycle.Lifecycle, log log.Logger) *Scheduler {
	quart := &Scheduler{
		scheduler: quartz.NewStdScheduler(),
		log:       log.WithName(Name),
	}
	quart.Start()
	m.BeforeStop(quart.Stop)
	return quart
}
