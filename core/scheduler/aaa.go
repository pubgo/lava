package scheduler

import (
	"context"
	stdLog "log"
	"os"
	"time"

	"github.com/pubgo/funk/stack"
	"github.com/pubgo/funk/v2/result"
)

var schedulerLog = stdLog.New(os.Stdout, "scheduler", stdLog.LstdFlags|stdLog.Lmsgprefix|stdLog.Lshortfile)

type JobExecutor interface {
	Name() string
	Exec(ctx context.Context, name string, metadata *JobMetadata) result.Result[[]byte]
}

var _ JobExecutor = (JobFunc)(nil)

type JobFunc func(ctx context.Context, name string, metadata *JobMetadata) result.Result[[]byte]

func (j JobFunc) Name() string {
	return stack.CallerWithFunc(j).String()
}

func (j JobFunc) Exec(ctx context.Context, name string, metadata *JobMetadata) result.Result[[]byte] {
	return j(ctx, name, metadata)
}

type JobRegister interface {
	RegisterSchedulerJob(reg JobRegistry)
}

type JobRegistry interface {
	Once(name string, delay time.Duration, fn JobFunc) result.Error
	Every(name string, dur time.Duration, fn JobFunc) result.Error
	Cron(name, expr string, fn JobFunc) result.Error
}

type JobManager interface {
	CreateJob(spec JobSpec) result.Error
	PatchJob(name string, config *JobConfig) result.Error
	PauseJob(name string) result.Error
	ResumeJob(name string) result.Error
	DeleteJob(name string) result.Error
	ReloadJob(name string) result.Error
	ListJobs() []Job
	GetJob(name string) Job
}

type JobExecT interface {
	string | JobFunc
}

type JobSpec struct {
	Name     string
	Config   *JobConfig
	Executor string
	Once     *OnceJob
	Ticker   *TickerJob
	Cron     *CronJob
}

type OnceJob struct {
	Delay time.Duration
}

type TickerJob struct {
	Dur time.Duration
}

type CronJob struct {
	Expr string
}

type Job struct {
	Spec     JobSpec
	Metadata JobMetadata
	ExecErr  error
	Status   Status
}

type Status string

const (
	StatusInit    Status = "init"
	StatusRunning Status = "running"
	StatusStop    Status = "stop"
)

type JobMetadata struct {
	Name          string
	Timeout       time.Duration
	MaxRetries    int
	RetryInterval time.Duration
	Replace       bool
	Location      *time.Location

	PreRunTime int64
	ExecTime   int64
}
