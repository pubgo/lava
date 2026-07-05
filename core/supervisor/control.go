package supervisor

import (
	"fmt"
	"time"
)

func (m *Manager) RestartServices() error {
	m.mu.RLock()
	names := make([]string, 0, len(m.services))
	for name := range m.services {
		names = append(names, name)
	}
	m.mu.RUnlock()

	for _, name := range names {
		if err := m.RestartService(name); err != nil {
			m.logger.Warn().Err(err).Str("name", name).Msg("failed to restart service")
		}
	}

	return nil
}

func (m *Manager) RestartService(name string) error {
	m.logger.Info().Str("name", name).Msg("restarting service in supervisor")

	if err := m.StopService(name); err != nil {
		return err
	}

	return m.StartService(name)
}

// StopService 暂停服务，但保留在 services map 中
func (m *Manager) StopService(name string) error {
	m.mu.Lock()
	srv := m.services[name]
	if srv == nil {
		m.mu.Unlock()
		m.logger.Warn().Str("name", name).Msg("service not found, cannot stop")
		return fmt.Errorf("%w, name=%s", ErrServiceNotFound, name)
	}

	if srv.stopped {
		m.mu.Unlock()
		m.logger.Warn().Str("name", name).Msg("service already stopped")
		return nil
	}

	cancel := srv.cancel
	done := srv.done
	srv.stopped = true
	m.mu.Unlock()

	if cancel != nil {
		cancel()
		if done != nil {
			<-done
		}
	}

	m.logger.Info().Str("name", name).Msg("stopped service in supervisor")
	return nil
}

// StartService 启动已暂停的服务
func (m *Manager) StartService(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	srv := m.services[name]
	if srv == nil {
		m.logger.Warn().Str("name", name).Msg("service not found, cannot start")
		return fmt.Errorf("%w, name=%s", ErrServiceNotFound, name)
	}

	if srv.running {
		m.logger.Warn().Str("name", name).Msg("service already running")
		return nil
	}

	srv.stopped = false
	srv.failed = false
	srv.consecFailures = 0
	srv.windowRestarts = 0
	srv.windowStart = time.Now()
	srv.currentDelay = srv.config.RestartDelay

	m.startRunner(srv)
	m.logger.Info().Str("name", name).Msg("started service in supervisor")
	return nil
}

// ResetService 重置失败的服务，清除所有重启计数
func (m *Manager) ResetService(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	srv := m.services[name]
	if srv == nil {
		return fmt.Errorf("%w, name=%s", ErrServiceNotFound, name)
	}

	srv.failed = false
	srv.restartCount = 0
	srv.consecFailures = 0
	srv.windowRestarts = 0
	srv.windowStart = time.Now()
	srv.currentDelay = srv.config.RestartDelay

	m.logger.Info().Str("name", name).Msg("reset service restart counters")
	return nil
}
