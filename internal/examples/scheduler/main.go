package main

import (
	"context"
	"fmt"
	"time"

	"github.com/pubgo/funk/v2/cmds/configcmd"
	"github.com/pubgo/funk/v2/cmds/envcmd"
	"github.com/pubgo/funk/v2/config"
	"github.com/pubgo/funk/v2/recovery"
	"github.com/pubgo/funk/v2/result"

	"github.com/pubgo/lava/v2/core/lavabuilder"
	"github.com/pubgo/lava/v2/core/logging"
	"github.com/pubgo/lava/v2/core/metrics"
	"github.com/pubgo/lava/v2/core/scheduler"
	"github.com/pubgo/lava/v2/servers/https"
)

type Config struct {
	metrics.MetricConfigLoader   `yaml:",inline"`
	logging.LogConfigLoader      `yaml:",inline"`
	https.HttpServerConfigLoader `yaml:",inline"`
}

var _ scheduler.JobRegister = (*schedulerExample)(nil)

type schedulerExample struct{}

func (s schedulerExample) RegisterSchedulerJob(reg scheduler.JobRegistry) {
	reg.Once("once_task", time.Second*10, func(ctx context.Context, name string, metadata *scheduler.JobMetadata) result.Result[[]byte] {
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

	builder := lavabuilder.New()
	builder.Provide(config.Load[Config])
	builder.Provide(envcmd.New)
	builder.Provide(configcmd.New[Config])
	builder.Provide(func() scheduler.JobRegister { return new(schedulerExample) })

	lavabuilder.Run(builder)
}
