package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/pubgo/funk/v2/buildinfo/version"
	"github.com/pubgo/funk/v2/config"
	"github.com/pubgo/funk/v2/debugs"
	"github.com/pubgo/funk/v2/env"
	"github.com/pubgo/funk/v2/log"
	"github.com/pubgo/funk/v2/recovery"
	"github.com/pubgo/funk/v2/result"
	"github.com/pubgo/funk/v2/result/resultchecker"

	"github.com/pubgo/lava/v2/cmds/configcmd"
	"github.com/pubgo/lava/v2/cmds/envcmd"
	"github.com/pubgo/lava/v2/core/lavabuilder"
	"github.com/pubgo/lava/v2/core/lifecycle"
	"github.com/pubgo/lava/v2/core/logging"
	"github.com/pubgo/lava/v2/core/metrics"
	"github.com/pubgo/lava/v2/core/scheduler"
	"github.com/pubgo/lava/v2/core/tunnel"
	_ "github.com/pubgo/lava/v2/core/tunnel/yamux"
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
		fmt.Println(debugs.Enabled.String())
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

	version.SetVersion("v1.0.0")
	version.SetProject("scheduler")
	config.SetConfigPath("internal/configs/scheduler.yaml")
	resultchecker.RegisterErrCheck(log.RecordErr())
	log.SetEnableChecker(func(ctx context.Context, lvl log.Level, name, message string, fields log.Fields) bool {
		//if lvl == zerolog.DebugLevel {
		//	return false
		//}
		return true
	})
	// debugs.SetEnabled()
	env.LoadFiles(".env").Must()

	builder := lavabuilder.New()
	builder.Provide(config.Load[Config])
	builder.Provide(envcmd.New)
	builder.Provide(configcmd.New[Config])
	builder.Provide(func() scheduler.JobRegister { return new(schedulerExample) })

	// Tunnel Agent 服务 - 使用 lifecycle hooks
	builder.Provide(func() lifecycle.Handler {
		return func(lc lifecycle.Lifecycle) {
			var agent tunnel.Agent

			lc.AfterStart(func(ctx context.Context) error {
				// 使用环境变量配置
				gatewayAddr := os.Getenv("TUNNEL_GATEWAY_ADDR")
				if gatewayAddr == "" {
					log.Info().Msg("Tunnel agent is disabled (TUNNEL_GATEWAY_ADDR not set)")
					return nil
				}

				serviceAddr := os.Getenv("TUNNEL_SERVICE_ADDR")
				if serviceAddr == "" {
					serviceAddr = ":8082"
				}

				log.Info().Str("gateway", gatewayAddr).Str("service_addr", serviceAddr).Msg("Starting tunnel agent")

				var err error
				agent, err = tunnel.NewAgentBuilder().
					WithGatewayAddr(gatewayAddr).
					WithServiceName("scheduler").
					WithServiceVersion("v1.0.0").
					AddEndpoint("http", serviceAddr, "/").
					AddEndpoint("debug", serviceAddr, "/debug").
					Build()
				if err != nil {
					return err
				}

				if err := agent.Start(ctx); err != nil {
					return err
				}

				log.Info().Msg("Tunnel agent started successfully")
				return nil
			})

			lc.BeforeStop(func(ctx context.Context) error {
				if agent != nil {
					return agent.Stop(ctx)
				}
				return nil
			})
		}
	})

	lavabuilder.Run(builder)
}
