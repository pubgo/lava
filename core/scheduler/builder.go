package scheduler

import (
	"github.com/pubgo/funk/log"
	"github.com/reugn/go-quartz/quartz"

	"github.com/pubgo/lava/core/lifecycle"
)

const Name = "scheduler"

func New(m lifecycle.Lifecycle, log log.Logger, opts []*Config) *Scheduler {
	quart := &Scheduler{
		scheduler: quartz.NewStdScheduler(),
		log:       log.WithName(Name),
	}

	quart.start()
	m.BeforeStop(quart.stop)
	return quart
}
