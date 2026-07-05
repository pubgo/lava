package supervisor

import (
	"context"
	"fmt"
)

// GetServicesInfo returns the status info of all services
func (m *Manager) GetServicesInfo() []*ServiceInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	services := make([]*ServiceInfo, 0, len(m.services))
	for _, srv := range m.services {
		services = append(services, m.serviceInfoFromRunner(srv))
	}
	return services
}

// GetServiceInfo returns the status info of a specific service
func (m *Manager) GetServiceInfo(name string) (*ServiceInfo, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	srv, ok := m.services[name]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrServiceNotFound, name)
	}

	return m.serviceInfoFromRunner(srv), nil
}

func (m *Manager) serviceInfoFromRunner(srv *serviceRunner) *ServiceInfo {
	metric := srv.service.Metric()
	metric.Status = m.statusFromRunner(srv)
	return &ServiceInfo{
		Metric:           metric,
		Stopped:          srv.stopped,
		Failed:           srv.failed,
		RestartCount:     srv.restartCount,
		ConsecFailures:   srv.consecFailures,
		WindowRestarts:   srv.windowRestarts,
		CurrentDelay:     srv.currentDelay,
		CurrentDelayStr:  srv.currentDelay.String(),
		WindowStart:      srv.windowStart,
		LastServiceStart: srv.lastServiceStart,
	}
}

func (m *Manager) statusFromRunner(runner *serviceRunner) ServiceStatus {
	if runner.failed {
		return StatusFailed
	}
	if runner.stopped {
		return StatusStopped
	}
	if runner.running && runner.consecFailures > 0 {
		return StatusCrashing
	}
	if runner.running {
		return StatusRunning
	}
	if runner.consecFailures > 0 {
		return StatusError
	}
	return StatusIdle
}

func (m *Manager) Has(name string) bool {
	m.mu.RLock()
	_, ok := m.services[name]
	m.mu.RUnlock()
	return ok
}

func (m *Manager) Add(srv Service, opts ...Option) error {
	config := DefaultServiceConfig()
	for _, opt := range opts {
		opt(&config)
	}
	return m.AddWithConfig(srv, config)
}

// AddWithConfig 添加服务并指定配置
func (m *Manager) AddWithConfig(srv Service, config ServiceConfig) error {
	config = normalizeServiceConfig(config)

	name := srv.Name()
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.services[name]; ok {
		return fmt.Errorf("%w, name=%s", ErrServiceAlreadyExists, name)
	}

	m.logger.Info().Str("name", name).Msg("add service to supervisor")
	runner := &serviceRunner{
		service: srv,
		config:  config,
		stopped: !config.AutoStart,
	}
	m.services[name] = runner

	if m.ctx != nil && !runner.stopped {
		m.startRunner(runner)
	}

	return nil
}

func (m *Manager) Delete(name string) error {
	m.mu.Lock()
	srv := m.services[name]
	if srv == nil {
		m.mu.Unlock()
		m.logger.Warn().Str("name", name).Msg("service not found, cannot delete")
		return fmt.Errorf("%w, name=%s", ErrServiceNotFound, name)
	}

	cancel := srv.cancel
	done := srv.done
	delete(m.services, name)
	m.mu.Unlock()

	if cancel != nil {
		cancel()
		if done != nil {
			<-done
		}
	}

	m.logger.Info().Str("name", name).Msg("delete service from supervisor")
	return nil
}

func (m *Manager) RemoveServices() error {
	type pendingStop struct {
		cancel context.CancelFunc
		done   chan struct{}
	}

	m.mu.Lock()
	pendings := make([]pendingStop, 0, len(m.services))
	for name, srv := range m.services {
		pendings = append(pendings, pendingStop{cancel: srv.cancel, done: srv.done})
		m.logger.Info().Str("name", name).Msg("removing service from supervisor")
	}
	m.services = make(map[string]*serviceRunner)
	m.mu.Unlock()

	for _, p := range pendings {
		if p.cancel != nil {
			p.cancel()
			if p.done != nil {
				<-p.done
			}
		}
	}

	return nil
}

func (m *Manager) Services() []Service {
	m.mu.RLock()
	defer m.mu.RUnlock()

	services := make([]Service, 0, len(m.services))
	for _, srv := range m.services {
		services = append(services, srv.service)
	}
	return services
}
