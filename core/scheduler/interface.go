package scheduler

import (
	"context"
	"time"
)

type JobFunc func(ctx context.Context, name string, metadata *JobMetadata) error

type JobRegister interface {
	RegisterSchedulerJob(reg JobRegistry)
}

type JobRegistry interface {
	Once(name string, delay time.Duration, fn JobFunc)
	Every(name string, dur time.Duration, fn JobFunc)
	Cron(name, expr string, fn JobFunc)
}
