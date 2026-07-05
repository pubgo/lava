package tunnelgateway

import (
	"bufio"
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/pubgo/funk/v2/log"

	"github.com/pubgo/lava/v2/core/tunnel"
)

func (g *tunnelGateway) healthCheckLoop() {
	defer g.wg.Done()

	interval := time.Duration(g.cfg.HealthCheckInterval) * time.Second
	if interval <= 0 {
		interval = 30 * time.Second
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-g.stopCh:
			return
		case <-ticker.C:
			g.checkServices()
		}
	}
}

func (g *tunnelGateway) checkServices() {
	heartbeatTimeout := time.Duration(g.cfg.HeartbeatTimeout) * time.Second
	if heartbeatTimeout <= 0 {
		heartbeatTimeout = 90 * time.Second
	}
	now := time.Now()

	g.mu.Lock()
	defer g.mu.Unlock()

	for name, svc := range g.services {
		if svc.session == nil || svc.session.IsClosed() {
			delete(g.services, name)
			log.Info().Str("service", name).Msg("Service removed (session closed)")
			continue
		}

		if !svc.info.LastHeartbeat.IsZero() && now.Sub(svc.info.LastHeartbeat) > heartbeatTimeout {
			svc.info.Status = tunnel.ServiceStatusOffline
			delete(g.services, name)
			log.Warn().
				Str("service", name).
				Dur("since_last_heartbeat", now.Sub(svc.info.LastHeartbeat)).
				Msg("Service removed (heartbeat timeout)")
			continue
		}

		status := g.checkServiceHealth(svc)
		if status != tunnel.ServiceStatusOnline {
			log.Warn().Str("service", name).Str("status", string(status)).Msg("Service health check failed")
		}
	}
}

func (g *tunnelGateway) checkServiceHealth(svc *registeredService) tunnel.ServiceStatus {
	if svc.session == nil || svc.session.IsClosed() {
		return tunnel.ServiceStatusOffline
	}

	allHealthy := true
	for i, endpoint := range svc.info.Endpoints {
		if svc.info.Endpoints[i].Metadata == nil {
			svc.info.Endpoints[i].Metadata = make(map[string]string)
		}
		if !g.checkEndpointHealth(svc, &endpoint) {
			allHealthy = false
			log.Warn().Str("service", svc.info.Name).Str("endpoint", string(endpoint.Type)).Str("address", endpoint.Address).Msg("Endpoint health check failed")
			svc.info.Endpoints[i].Metadata["health_status"] = "unhealthy"
		} else {
			svc.info.Endpoints[i].Metadata["health_status"] = "healthy"
		}
	}

	if allHealthy {
		svc.info.Status = tunnel.ServiceStatusOnline
		return tunnel.ServiceStatusOnline
	}
	svc.info.Status = tunnel.ServiceStatusUnhealthy
	return tunnel.ServiceStatusUnhealthy
}

func (g *tunnelGateway) checkEndpointHealth(svc *registeredService, endpoint *tunnel.Endpoint) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	switch endpoint.Type {
	case tunnel.EndpointTypeHTTP:
		return g.checkHTTPEndpointHealth(ctx, svc, endpoint)
	case tunnel.EndpointTypeGRPC:
		return g.checkGRPCEndpointHealth(ctx, svc, endpoint)
	case tunnel.EndpointTypeDebug:
		return g.checkDebugEndpointHealth(ctx, svc, endpoint)
	default:
		return true
	}
}

func (g *tunnelGateway) checkHTTPEndpointHealth(ctx context.Context, svc *registeredService, endpoint *tunnel.Endpoint) bool {
	stream, err := svc.session.Open(ctx)
	if err != nil {
		return false
	}
	defer func() {
		if err := stream.Close(); err != nil {
			log.Warn().Err(err).Msg("Gateway: failed to close health check stream")
		}
	}()

	meta := tunnel.RequestMeta{
		ServiceID:    svc.info.ID,
		EndpointType: tunnel.EndpointTypeHTTP,
		Path:         endpoint.Path + "/health",
		Method:       "GET",
	}
	payload, err := json.Marshal(meta)
	if err != nil {
		log.Warn().Err(err).Str("service", svc.info.Name).Msg("Gateway: failed to marshal health check meta")
		return false
	}

	msg := &tunnel.Message{
		Type:    tunnel.MessageTypeHTTPRequest,
		Payload: payload,
	}
	if err := g.sendMessage(stream, msg); err != nil {
		return false
	}

	req, err := http.NewRequestWithContext(ctx, "GET", endpoint.Address+"/health", nil)
	if err != nil {
		return false
	}
	if err := req.Write(stream); err != nil {
		return false
	}

	resp, err := http.ReadResponse(bufio.NewReader(stream), req)
	if err != nil {
		return false
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Warn().Err(err).Msg("Gateway: failed to close health check response body")
		}
	}()

	return resp.StatusCode == http.StatusOK
}

func (g *tunnelGateway) checkGRPCEndpointHealth(ctx context.Context, svc *registeredService, endpoint *tunnel.Endpoint) bool {
	return true
}

func (g *tunnelGateway) checkDebugEndpointHealth(ctx context.Context, svc *registeredService, endpoint *tunnel.Endpoint) bool {
	stream, err := svc.session.Open(ctx)
	if err != nil {
		return false
	}
	if err := stream.Close(); err != nil {
		log.Warn().Err(err).Msg("Gateway: failed to close debug health check stream")
		return false
	}
	return true
}
