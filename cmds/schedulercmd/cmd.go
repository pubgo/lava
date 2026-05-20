package schedulercmd

import (
	"context"
	"os"

	"github.com/pubgo/dix/v2"
	"github.com/pubgo/funk/v2/assert"
	"github.com/pubgo/funk/v2/buildinfo/version"
	"github.com/pubgo/funk/v2/log"
	"github.com/pubgo/redant"

	"github.com/pubgo/lava/v2/core/lifecycle"
	"github.com/pubgo/lava/v2/core/running"
	"github.com/pubgo/lava/v2/core/scheduler/schedulerbuilder"
	"github.com/pubgo/lava/v2/core/supervisor"
	supervisordebug "github.com/pubgo/lava/v2/core/supervisor/debug"
	"github.com/pubgo/lava/v2/core/tunnel"
	"github.com/pubgo/lava/v2/core/tunnel/tunnelagent"
	"github.com/pubgo/lava/v2/core/tunnel/tunneldebug"
	"github.com/pubgo/lava/v2/pkg/cliutil"
	"github.com/pubgo/lava/v2/servers/https"
)

func New(di *dix.Dix) *redant.Command {
	return &redant.Command{
		Use:   "cron",
		Short: cliutil.UsageDesc("crontab scheduler service %s(%s)", version.Project(), version.Version()),
		Handler: func(ctx context.Context, i *redant.Invocation) error {
			di.Provide(schedulerbuilder.NewService)
			di.Provide(https.New)
			params := dix.Inject(di, new(struct {
				LC       lifecycle.Getter
				Services []supervisor.Service `dix:"scheduler"`
			}))

			manager := supervisor.Default(params.LC)
			supervisordebug.Register(manager)
			for _, svc := range params.Services {
				assert.Exit(manager.Add(svc))
			}

			// 集成 Tunnel Agent
			// Agent 主动连接 Gateway，将本服务的 HTTP 和 Debug 端点暴露出去
			gatewayAddr := os.Getenv("TUNNEL_GATEWAY_ADDR")
			if gatewayAddr == "" {
				gatewayAddr = "localhost:7007" // 默认 Gateway 地址
			}

			// 获取本地服务地址（通过环境变量配置）
			httpAddr := os.Getenv("HTTP_ADDR")
			if httpAddr == "" {
				httpAddr = ":" + running.HttpPort.String()
			}
			debugAddr := os.Getenv("DEBUG_ADDR")
			if debugAddr == "" {
				debugAddr = ":" + running.HttpPort.String()
			}

			// 获取服务名，优先使用环境变量，其次使用 buildinfo，最后使用默认值
			serviceName := os.Getenv("SERVICE_NAME")
			if serviceName == "" {
				serviceName = version.Project()
			}
			if serviceName == "" {
				serviceName = "scheduler"
			}

			serviceVersion := version.Version()
			if serviceVersion == "" {
				serviceVersion = "dev"
			}

			agent := tunnelagent.NewAgent(&tunnel.AgentConfig{
				GatewayAddr:    gatewayAddr,
				ServiceName:    serviceName,
				ServiceVersion: serviceVersion,
				Metadata: map[string]string{
					"instance": os.Getenv("HOSTNAME"),
				},

				Endpoints: []tunnel.EndpointConfig{
					{Type: "http", LocalAddr: httpAddr, Path: "/"},
					{Type: "debug", LocalAddr: debugAddr, Path: "/debug"},
				},
			})

			err := agent.Start(ctx)
			if err != nil {
				log.Error().Err(err).Msg("Failed to start tunnel agent")
			} else {
				// 注册到 tunneldebug，可以在 /debug/tunnel 查看 Agent 状态
				tunneldebug.SetAgent(agent)
				assert.Exit(manager.Add(&tunnelAgentService{agent: agent}))
				log.Info().
					Str("gateway", gatewayAddr).
					Str("service", version.Project()).
					Msg("Tunnel Agent integrated")
			}

			return manager.Run(ctx)
		},
	}
}

// tunnelAgentService 包装 Agent 为 supervisor.Service
type tunnelAgentService struct {
	agent tunnel.Agent
	err   error
}

func (s *tunnelAgentService) Name() string {
	return "tunnel-agent"
}

func (s *tunnelAgentService) Error() error {
	return s.err
}

func (s *tunnelAgentService) String() string {
	return "Tunnel Agent Service - connects to gateway and exposes local services"
}

func (s *tunnelAgentService) Serve(ctx context.Context) error {
	log.Info().Msg("Starting Tunnel Agent...")
	if err := s.agent.Start(ctx); err != nil {
		s.err = err
		return err
	}

	// 等待上下文取消
	<-ctx.Done()

	log.Info().Msg("Stopping Tunnel Agent...")
	return s.agent.Stop(context.Background())
}

func (s *tunnelAgentService) Metric() *supervisor.Metric {
	return &supervisor.Metric{
		Name:   s.Name(),
		Status: supervisor.StatusRunning,
	}
}
