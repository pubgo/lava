package tunnelcmd

import (
	"context"
	"os"

	"github.com/pubgo/dix/v2"
	"github.com/pubgo/funk/v2/assert"
	"github.com/pubgo/funk/v2/buildinfo/version"
	"github.com/pubgo/funk/v2/log"
	"github.com/pubgo/redant"

	"github.com/pubgo/lava/v2/core/lifecycle"
	"github.com/pubgo/lava/v2/core/p2p"
	"github.com/pubgo/lava/v2/core/p2p/p2pbuilder"
	_ "github.com/pubgo/lava/v2/core/p2p/transport"
	"github.com/pubgo/lava/v2/core/supervisor"
	supervisordebug "github.com/pubgo/lava/v2/core/supervisor/debug"
	"github.com/pubgo/lava/v2/core/tunnel"
	"github.com/pubgo/lava/v2/core/tunnel/tunnelagent"
	"github.com/pubgo/lava/v2/core/tunnel/tunneldebug"
	"github.com/pubgo/lava/v2/pkg/cliutil"
)

func newAgentCommand(di *dix.Dix) *redant.Command {
	return &redant.Command{
		Use:   "agent",
		Short: cliutil.UsageDesc("tunnel agent with optional P2P %s(%s)", version.Project(), version.Version()),
		Handler: func(ctx context.Context, i *redant.Invocation) error {
			gatewayAddr := envOr("TUNNEL_GATEWAY_ADDR", "localhost:7007")
			authToken := p2p.AuthTokenFromEnv()
			serviceName := envOr("SERVICE_NAME", version.Project())
			if serviceName == "" {
				serviceName = "tunnel-agent"
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
				}, authToken),
			}
			agentCfg.TLS.ApplyEnv()
			agent := tunnelagent.New(agentCfg)

			params := dix.Inject(di, new(struct {
				LC       lifecycle.Lifecycle
				LCGetter lifecycle.Getter
				Services []supervisor.Service
			}))

			manager := supervisor.Default(params.LCGetter)
			supervisordebug.Register(manager)
			for _, svc := range params.Services {
				assert.Exit(manager.Add(svc))
			}

			tunneldebug.SetAgent(agent)
			assert.Exit(manager.Add(&tunnelAgentService{agent: agent}))

			peerID := p2p.PeerIDFromEnv()
			if peerID != "" {
				cfg := p2p.ConfigFromEnv()
				if cfg.AuthToken == "" && authToken != "" {
					cfg.AuthToken = authToken
				}
				_, err := p2pbuilder.New(p2pbuilder.Params{
					LC:            params.LC,
					Agent:         agent,
					PeerID:        peerID,
					ListenOnStart: true,
					Config:        &cfg,
				})
				if err != nil {
					return err
				}
				log.Info().Str("peer_id", peerID).Msg("P2P enabled on tunnel agent")
			}

			debugAddr := envOr("TUNNEL_ADMIN_ADDR", ":6067")
			assert.Exit(manager.Add(newDebugServer(debugAddr)))

			log.Info().
				Str("gateway", gatewayAddr).
				Str("service", serviceName).
				Str("peer_id", peerID).
				Str("admin_addr", debugAddr).
				Bool("p2p_enabled", peerID != "").
				Bool("auth_enabled", authToken != "").
				Msg("Starting Tunnel Agent")

			return manager.Run(ctx)
		},
	}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

type tunnelAgentService struct {
	agent tunnel.Agent
	err   error
}

func (s *tunnelAgentService) Name() string { return "tunnel-agent" }

func (s *tunnelAgentService) Error() error { return s.err }

func (s *tunnelAgentService) String() string {
	return "Tunnel Agent Service - connects to gateway"
}

func (s *tunnelAgentService) Serve(ctx context.Context) error {
	log.Info().Msg("Starting Tunnel Agent...")
	if err := s.agent.Start(ctx); err != nil {
		s.err = err
		return err
	}
	<-ctx.Done()
	log.Info().Msg("Stopping Tunnel Agent...")
	return s.agent.Stop(context.Background())
}

func (s *tunnelAgentService) Metric() *supervisor.Metric {
	m := &supervisor.Metric{Name: s.Name()}
	if s.err != nil {
		m.Status = supervisor.StatusError
		m.LastError = s.err.Error()
		return m
	}
	switch s.agent.Status() {
	case tunnel.StatusConnected:
		m.Status = supervisor.StatusRunning
	case tunnel.StatusConnecting, tunnel.StatusReconnecting:
		m.Status = supervisor.StatusCrashing
	default:
		m.Status = supervisor.StatusStopped
	}
	return m
}
