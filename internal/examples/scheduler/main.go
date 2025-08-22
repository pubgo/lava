package main

import (
	"context"
	"fmt"
	"time"

	"github.com/pubgo/dix/dixinternal"
	"github.com/pubgo/funk/cmds/configcmd"
	"github.com/pubgo/funk/cmds/envcmd"
	"github.com/pubgo/funk/config"
	"github.com/pubgo/funk/log"
	"github.com/pubgo/funk/recovery"
	"github.com/pubgo/funk/v2/result"
	"github.com/pubgo/lava/core/lavabuilder"
	"github.com/pubgo/lava/core/logging"
	"github.com/pubgo/lava/core/metrics"
	"github.com/pubgo/lava/core/scheduler"
	"github.com/pubgo/lava/servers/https"
	"github.com/rs/zerolog"
)

type Config struct {
	metrics.MetricConfigLoader   `yaml:",inline"`
	logging.LogConfigLoader      `yaml:",inline"`
	https.HttpServerConfigLoader `yaml:",inline"`
}

var _ scheduler.JobRegister = (*schedulerExample)(nil)

type schedulerExample struct {
}

func (s schedulerExample) RegisterSchedulerJob(reg scheduler.JobRegistry) {
	reg.Once("once_task", time.Second*12, func(ctx context.Context, name string, metadata *scheduler.JobMetadata) result.Result[[]byte] {
		fmt.Printf("exec once task: %s: %#v\n", name, metadata)
		time.Sleep(time.Second * 5)
		return result.OK([]byte("once"))
	})

	reg.Every("every_task", time.Second*5, func(ctx context.Context, name string, metadata *scheduler.JobMetadata) result.Result[[]byte] {
		fmt.Printf("exec every task: %s: %#v\n", name, metadata)
		time.Sleep(time.Second * 1)
		return result.OK([]byte("every"))
	})

	reg.Cron("cron_task", "*/7 * * * * *", func(ctx context.Context, name string, metadata *scheduler.JobMetadata) result.Result[[]byte] {
		fmt.Printf("exec cron task: %s: %#v\n", name, metadata)
		time.Sleep(time.Second * 2)
		return result.OK([]byte("cron"))
	})
}

func main() {
	defer recovery.Exit()

	dixinternal.SetLog(func(logger log.Logger) log.Logger {
		return logger.WithLevel(zerolog.InfoLevel)
	})

	builder := lavabuilder.New()
	builder.Provide(config.Load[Config])
	builder.Provide(envcmd.New)
	builder.Provide(configcmd.New[Config])
	builder.Provide(func() scheduler.JobRegister {
		return &schedulerExample{}
	})

	lavabuilder.Run(builder)
}
