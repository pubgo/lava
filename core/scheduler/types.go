package scheduler

import (
	"context"
	"sync"
	"time"

	"github.com/pubgo/funk/v2/clone"
	"github.com/pubgo/funk/v2/result"
	"github.com/pubgo/funk/v2/stack"
	"github.com/reugn/go-quartz/quartz"
	"go.uber.org/atomic"
)

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
	ListJobs() []*Job
	GetJob(name string) result.Result[*Job]
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
	Spec     *JobSpec
	Metadata JobMetadata
	Status   Status

	PreExecTime int64
	ExecTime    int64

	Error  error
	Result []byte
	Runs   uint64
}

type jobTask struct {
	spec     *JobSpec
	executor JobExecutor

	trigger *triggerImpl
	runs    atomic.Uint64
	jobKey  *quartz.JobKey
	status  Status

	resultMu sync.RWMutex
	result   result.Result[[]byte]
}

func (job *jobTask) ToJob() *Job {
	preExecTime, execTime, _ := job.trigger.Snapshot()
	res := job.getResult()
	resultData := res.UnwrapOrEmpty()
	if resultData != nil {
		resultData = append([]byte(nil), resultData...)
	}

	return &Job{
		Status:      job.status,
		PreExecTime: preExecTime,
		ExecTime:    execTime,
		Error:       res.GetErr(),
		Result:      resultData,
		Runs:        job.runs.Load(),
		Spec:        clone.Clone(job.spec),
	}
}

func (job *jobTask) setResult(res result.Result[[]byte]) {
	job.resultMu.Lock()
	job.result = res
	job.resultMu.Unlock()
}

func (job *jobTask) getResult() result.Result[[]byte] {
	job.resultMu.RLock()
	defer job.resultMu.RUnlock()
	return job.result
}

type Status string

const (
	StatusRunning Status = "running"
	StatusStop    Status = "stop"
)

type JobMetadata struct {
	Name          string
	Timeout       time.Duration
	MaxRetries    int
	RetryInterval time.Duration
	Replace       bool
	Location      string

	NextExecTime int64
	ExecTime     int64
}
