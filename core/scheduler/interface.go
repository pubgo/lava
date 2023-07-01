package scheduler

import "context"

type JobFunc func(ctx context.Context, name string) error

type CronRouter interface {
	Crontab(s *Scheduler)
}
