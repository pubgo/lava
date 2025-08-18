package scheduler

import (
	"context"
	"time"

	"github.com/reugn/go-quartz/quartz"
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

type JobManager interface {
	Add(spec AddJobSpec)
	Pause(name string)
	Resume(name string)
	Delete(name string)
	Reload(name string)
	List()
	Get(name string)
}

type AddJobSpec struct {
	Name    string
	Setting JobSetting
	Job     JobFunc

	Once  *OnceJob
	Every *EveryJob
	Cron  *CronJob
}

type OnceJob struct {
	Delay time.Duration
}

type EveryJob struct {
	Dur time.Duration
}

type CronJob struct {
	Expr string
}

type Job struct {
	Name     string
	Delay    *time.Duration
	Dur      *time.Duration
	CronExpr string
	Metadata JobMetadata
	ExecErr  error
	JobKey   *quartz.JobKey
	Location *time.Location
}
