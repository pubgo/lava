package tunnelgateway

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/pubgo/funk/v2/log"

	"github.com/pubgo/lava/v2/core/tunnel"
)

func (g *tunnelGateway) Start(ctx context.Context) error {
	if g.running.Load() {
		return tunnel.ErrGatewayAlreadyRunning
	}

	transport, err := tunnel.NewTransport(g.cfg.Transport, g.cfg.TransportOptions)
	if err != nil {
		return err
	}
	g.transport = transport

	listener, err := transport.Listen(ctx, g.cfg.ListenAddr)
	if err != nil {
		return err
	}
	g.listener = listener

	g.stopCh = make(chan struct{})
	g.running.Store(true)
	g.status = tunnel.GatewayStatusRunning

	g.wg.Add(1)
	go g.acceptLoop()

	g.wg.Add(1)
	go g.healthCheckLoop()

	if g.cfg.HTTPPort > 0 {
		g.wg.Add(1)
		go g.startHTTPProxy()
	}

	if g.cfg.DebugPort > 0 {
		g.wg.Add(1)
		go g.startDebugProxy()
	}

	if g.cfg.GRPCPort > 0 {
		g.wg.Add(1)
		go g.startGRPCProxy()
	}

	log.Info().
		Str("tunnel_addr", g.cfg.ListenAddr).
		Int("http_port", g.cfg.HTTPPort).
		Int("grpc_port", g.cfg.GRPCPort).
		Int("debug_port", g.cfg.DebugPort).
		Msg("Gateway started, waiting for agents to connect...")

	return nil
}

func (g *tunnelGateway) Stop(ctx context.Context) error {
	if !g.running.Load() {
		return nil
	}

	g.stopOnce.Do(func() {
		close(g.stopCh)
	})

	shutdownCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	if g.httpServer != nil {
		if err := g.httpServer.Shutdown(shutdownCtx); err != nil {
			log.Warn().Err(err).Msg("Gateway: failed to shutdown HTTP server")
		}
	}
	if g.debugServer != nil {
		if err := g.debugServer.Shutdown(shutdownCtx); err != nil {
			log.Warn().Err(err).Msg("Gateway: failed to shutdown debug server")
		}
	}
	if g.grpcListener != nil {
		if err := g.grpcListener.Close(); err != nil {
			log.Warn().Err(err).Msg("Gateway: failed to close GRPC listener")
		}
	}

	if g.listener != nil {
		if err := g.listener.Close(); err != nil {
			log.Warn().Err(err).Msg("Gateway: failed to close listener")
		}
	}

	g.mu.Lock()
	for _, svc := range g.services {
		if svc.session != nil {
			if err := svc.session.Close(); err != nil {
				log.Warn().Err(err).Str("service", svc.info.Name).Msg("Gateway: failed to close session")
			}
		}
	}
	g.services = make(map[string]*registeredService)
	g.mu.Unlock()

	done := make(chan struct{})
	go func() {
		g.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(3 * time.Second):
		log.Warn().Msg("Gateway stop timeout, force closing")
	}

	g.running.Store(false)
	g.status = tunnel.GatewayStatusStopped
	return nil
}

func (g *tunnelGateway) Services() []*tunnel.ServiceInfo {
	g.mu.RLock()
	defer g.mu.RUnlock()

	services := make([]*tunnel.ServiceInfo, 0, len(g.services))
	for _, svc := range g.services {
		services = append(services, svc.info)
	}
	return services
}

func (g *tunnelGateway) GetService(name string) (*tunnel.ServiceInfo, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	svc, ok := g.services[name]
	if !ok {
		return nil, tunnel.ErrServiceNotFound
	}
	return svc.info, nil
}

func (g *tunnelGateway) Status() tunnel.GatewayStatus {
	return g.status
}

// ServeHTTP implements http.Handler for the gateway status
func (g *tunnelGateway) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	services := g.Services()

	status := map[string]any{
		"status":       g.status.String(),
		"listen_addr":  g.cfg.ListenAddr,
		"transport":    g.cfg.Transport,
		"num_services": len(services),
		"services":     services,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(status); err != nil {
		log.Warn().Err(err).Msg("Gateway: failed to encode status")
	}
}
