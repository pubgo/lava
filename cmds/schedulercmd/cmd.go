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
	_ "github.com/pubgo/lava/v2/core/tunnel/yamux" // 注册 yamux 传输
	"github.com/pubgo/lava/v2/pkg/cliutil"
	"github.com/pubgo/lava/v2/servers/https"
)

func New(di *dix.Dix) *redant.Command {
	return &redant.Command{
		Use:   "cron",
		Short: cliutil.UsageDesc("crontab scheduler service %s(%s)", version.Project(), version.Version()),
		Handler: func(ctx context.Context, i *redant.Invocation) error {
			di.Provide(schedulerbuilder.NewService)
			params := dix.Inject(di, new(struct {
				LC         lifecycle.Getter
				Scheduler  schedulerbuilder.ResponseParams
				HTTPParams https.Params
			}))

			manager := supervisor.Default(params.LC)
			supervisordebug.Register(manager)
			assert.Exit(manager.Add(params.Scheduler.Service))
			assert.Exit(manager.Add(https.New(params.HTTPParams)))

			// 集成 Tunnel Agent
			// Agent 主动连接 Gateway，将本服务的 HTTP 和 Debug 端点暴露出去
			gatewayAddr := os.Getenv("TUNNEL_GATEWAY_ADDR")
			if gatewayAddr == "" {
				gatewayAddr = "localhost:7007" // 默认 Gateway 地址
			}

			// 获取本地服务地址（通过环境变量配置）
			httpAddr := running.HttpListenAddr()
			debugAddr := running.DebugListenAddr()

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

			agentCfg := &tunnel.AgentConfig{
				GatewayAddr:    gatewayAddr,
				Transport:      tunnel.TransportYamux,
				ServiceName:    serviceName,
				ServiceVersion: serviceVersion,
				Metadata: tunnel.ApplyAuthTokenMetadata(map[string]string{
					"instance": os.Getenv("HOSTNAME"),
				}, os.Getenv("TUNNEL_AUTH_TOKEN")),
				Endpoints: []tunnel.EndpointConfig{
					{Type: "http", LocalAddr: httpAddr, Path: "/"},
					{Type: "debug", LocalAddr: debugAddr, Path: "/debug"},
				},
			}
			agentCfg.TLS.ApplyEnv()
			agent := tunnelagent.NewAgent(agentCfg)

			// 注册到 tunneldebug，可以在 /debug/tunnel 查看 Agent 状态
			// 实际启动交由 supervisor 生命周期统一管理，避免重复 Start
			tunneldebug.SetAgent(agent)
			assert.Exit(manager.Add(tunnelagent.NewSupervisorService(agent)))
			log.Info().
				Str("gateway", gatewayAddr).
				Str("service", serviceName).
				Msg("Tunnel Agent integrated")

			return manager.Run(ctx)
		},
	}
}
