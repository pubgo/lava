package tunnelagent

import (
	"context"

	"github.com/pubgo/funk/v2/log"

	"github.com/pubgo/lava/v2/core/supervisor"
	"github.com/pubgo/lava/v2/core/tunnel"
)

// SupervisorService wraps a tunnel.Agent as a supervisor.Service.
type SupervisorService struct {
	agent tunnel.Agent
	err   error
}

// NewSupervisorService returns a supervisor.Service for the given agent.
func NewSupervisorService(agent tunnel.Agent) *SupervisorService {
	return &SupervisorService{agent: agent}
}

func (s *SupervisorService) Name() string { return "tunnel-agent" }

func (s *SupervisorService) Error() error { return s.err }

func (s *SupervisorService) String() string {
	return "Tunnel Agent Service - connects to gateway and exposes local services"
}

func (s *SupervisorService) Serve(ctx context.Context) error {
	log.Info().Msg("Starting Tunnel Agent...")
	if err := s.agent.Start(ctx); err != nil {
		s.err = err
		return err
	}
	<-ctx.Done()
	log.Info().Msg("Stopping Tunnel Agent...")
	return s.agent.Stop(context.Background())
}

func (s *SupervisorService) Metric() *supervisor.Metric {
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

var _ supervisor.Service = (*SupervisorService)(nil)
