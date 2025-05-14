package scheduler

import (
	"context"
	"time"
)

type JobFunc func(ctx context.Context, name string) error

type Register interface {
	RegisterCrontabScheduler(reg Registry)
}

type Registry interface {
	Once(name string, delay time.Duration, fn JobFunc)
	Every(name string, dur time.Duration, fn JobFunc)
	Cron(name, expr string, fn JobFunc)
}
