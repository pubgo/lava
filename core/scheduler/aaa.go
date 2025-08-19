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
	Once(name string, delay time.Duration, fn JobFunc)
	Every(name string, dur time.Duration, fn JobFunc)
	Cron(name, expr string, fn JobFunc)
}

type JobManager interface {
	Create(spec AddJobSpec) result.Error
	Patch(name string, config JobConfig) error
	Pause(name string) error
	Resume(name string) error
	Delete(name string) error
	Reload(name string) error
	List() []Job
	Get(name string) Job
}

type AddJobSpec struct {
	Name     string
	Config   JobConfig
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
	Name     string
	Metadata JobMetadata
	ExecErr  error
	Executor string
	Once     *OnceJob
	Ticker   *TickerJob
	Cron     *CronJob
	Status   Status
}

type Status string

const (
	StatusInit    Status = "init"
	StatusRunning Status = "running"
	StatusStop    Status = "stop"
)
