package tunnelgateway

import (
	"encoding/json"
	"time"

	"github.com/pubgo/funk/v2/log"

	"github.com/pubgo/lava/v2/core/tunnel"
)

func (g *tunnelGateway) acceptLoop() {
	defer g.wg.Done()

	for {
		select {
		case <-g.stopCh:
			return
		default:
		}

		session, err := g.listener.Accept()
		if err != nil {
			select {
			case <-g.stopCh:
				return
			default:
				log.Warn().Err(err).Msg("Failed to accept connection")
				continue
			}
		}

		go g.handleSession(session)
	}
}

func (g *tunnelGateway) handleSession(session tunnel.Session) {
	agentID := session.RemoteAddr().String()
	log.Info().Str("agent", agentID).Msg("Agent connected")

	for {
		select {
		case <-g.stopCh:
			return
		default:
		}

		if session.IsClosed() {
			g.removeAgentServices(agentID)
			return
		}

		stream, err := session.Accept()
		if err != nil {
			if session.IsClosed() {
				g.removeAgentServices(agentID)
				return
			}
			log.Warn().Err(err).Msg("Failed to accept stream")
			continue
		}

		go g.handleStream(agentID, session, stream)
	}
}

func (g *tunnelGateway) handleStream(agentID string, session tunnel.Session, stream tunnel.Stream) {
	defer func() {
		if err := stream.Close(); err != nil {
			log.Warn().Err(err).Str("agent", agentID).Msg("Gateway: failed to close stream")
		}
	}()

	msg, err := tunnel.ReadMessage(stream)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to read message")
		return
	}

	switch msg.Type {
	case tunnel.MessageTypeRegister:
		g.handleRegister(agentID, session, msg)
	case tunnel.MessageTypeDeregister:
		g.handleDeregister(agentID, msg)
	case tunnel.MessageTypeHeartbeat:
		g.handleHeartbeat(agentID)
	case tunnel.MessageTypeP2PRegister:
		g.handleP2PRegister(agentID, session, msg)
	case tunnel.MessageTypeP2PSignal:
		g.handleP2PSignal(agentID, session, msg)
	default:
		log.Warn().Uint8("type", uint8(msg.Type)).Msg("Unknown message type")
	}
}

func (g *tunnelGateway) handleRegister(agentID string, session tunnel.Session, msg *tunnel.Message) {
	if len(msg.Payload) == 0 {
		log.Warn().Str("agent", agentID).Msg("Register: empty payload")
		return
	}

	var service tunnel.ServiceInfo
	if err := json.Unmarshal(msg.Payload, &service); err != nil {
		log.Warn().Err(err).Str("agent", agentID).Msg("Register: failed to parse service info")
		return
	}

	if g.authProvider != nil {
		if err := g.authProvider.Authenticate(&service); err != nil {
			log.Warn().Err(err).Str("agent", agentID).Str("service", service.Name).Msg("Register: authentication failed")
			return
		}
	}

	now := time.Now()
	service.RegisterTime = now
	service.LastHeartbeat = now
	service.Status = tunnel.ServiceStatusOnline

	g.mu.Lock()
	g.services[service.Name] = &registeredService{
		info:    &service,
		session: session,
		agent:   agentID,
	}
	registered := len(g.services)
	g.mu.Unlock()

	if g.metrics != nil {
		g.metrics.ObserveRegister(service.Name)
		g.metrics.SetRegisteredServices(registered)
	}

	log.Info().Str("service", service.Name).Str("agent", agentID).Msg("Service registered")
}

func (g *tunnelGateway) handleDeregister(agentID string, msg *tunnel.Message) {
	if len(msg.Payload) == 0 {
		return
	}

	serviceName := string(msg.Payload)

	g.mu.Lock()
	defer g.mu.Unlock()

	svc, ok := g.services[serviceName]
	if !ok {
		return
	}
	if svc.agent != agentID {
		log.Warn().
			Str("service", serviceName).
			Str("agent", agentID).
			Str("owner", svc.agent).
			Msg("Deregister: rejected (agent mismatch)")
		return
	}
	delete(g.services, serviceName)

	log.Info().Str("service", serviceName).Str("agent", agentID).Msg("Service deregistered")
}

func (g *tunnelGateway) handleHeartbeat(agentID string) {
	now := time.Now()
	g.mu.Lock()
	for _, svc := range g.services {
		if svc.agent == agentID {
			svc.info.LastHeartbeat = now
		}
	}
	g.mu.Unlock()

	log.Debug().Str("agent", agentID).Msg("Heartbeat received")
}

func (g *tunnelGateway) removeAgentServices(agentID string) {
	g.mu.Lock()
	defer g.mu.Unlock()

	for name, svc := range g.services {
		if svc.agent == agentID {
			delete(g.services, name)
			log.Info().Str("service", name).Str("agent", agentID).Msg("Service removed (agent disconnected)")
		}
	}
	g.removePeerLocked(agentID)
}
