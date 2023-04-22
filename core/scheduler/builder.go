package scheduler

import (
	"context"
	"fmt"

	"github.com/pubgo/funk/log"
	"github.com/reugn/go-quartz/quartz"

	"github.com/pubgo/lava/core/lifecycle"
)

const Name = "scheduler"

func New(m lifecycle.Lifecycle, log log.Logger, opts []*Config) *Scheduler {
	var config = make(map[string]JobSetting)
	if len(opts) > 0 && opts[0] != nil {
		for _, opt := range *opts[0] {
			if _, ok := config[opt.Name]; ok {
				panic(fmt.Sprintf("schedule job(%s) exists", opt.Name))
			}

			config[opt.Name] = opt
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	quart := &Scheduler{
		config:    config,
		scheduler: quartz.NewStdScheduler(),
		log:       log.WithName(Name),
		ctx:       ctx,
		cancel:    cancel,
	}

	quart.start()
	m.BeforeStop(quart.stop)
	return quart
}
