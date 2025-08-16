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
	Once(name string, delay time.Duration, fn JobFunc, opts ...Options)
	Every(name string, dur time.Duration, fn JobFunc, opts ...Options)
	Cron(name, expr string, fn JobFunc, opts ...Options)
}

type Options struct {
}
