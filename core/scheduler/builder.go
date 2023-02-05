package scheduler

import (
	"github.com/pubgo/funk/lifecycle"
	"github.com/pubgo/funk/log"
	"github.com/reugn/go-quartz/quartz"
)

const Name = "scheduler"

func New(m lifecycle.Lifecycle, log log.Logger) *Scheduler {
	var quart = &Scheduler{
		scheduler: quartz.NewStdScheduler(),
		log:       log.WithName(Name),
	}
	quart.Start()
	m.BeforeStop(quart.Stop)
	return quart
}
