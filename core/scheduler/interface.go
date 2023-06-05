package scheduler

type CronRouter interface {
	Crontab(s *Scheduler)
}
