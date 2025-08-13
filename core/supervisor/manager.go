package supervisor

import (
	"context"

	"github.com/pubgo/funk/errors"
	"github.com/pubgo/funk/errors/errcheck"
	"github.com/pubgo/funk/log"
	"github.com/pubgo/funk/running"
	"github.com/thejerf/suture/v4"
)

type serviceWrapper struct {
	token   suture.ServiceToken
	service Service
}

func DefaultManager() *Manager {
	return NewManager(running.Project)
}

func NewManager(name string) *Manager {
	return &Manager{
		supervisor: suture.New(name, SpecWithInfoLogger()),
		services:   make(map[string]*serviceWrapper),
		logger:     log.GetLogger(name),
	}
}

type Manager struct {
	logger     log.Logger
	supervisor *Supervisor
	services   map[string]*serviceWrapper
}

func (m *Manager) Has(name string) bool {
	_, ok := m.services[name]
	return ok
}

func (m *Manager) OnClose(fn func()) {
	m.supervisor.Add(doneService(fn))
}

func (m *Manager) Add(srv Service) error {
	name := srv.Name()
	if _, ok := m.services[name]; ok {
		return errors.Errorf("service already exists, name=%s", name)
	}

	m.logger.Info().Str("name", name).Msg("add service to supervisor")
	m.services[name] = &serviceWrapper{service: srv, token: m.supervisor.Add(srv)}
	return nil
}

func (m *Manager) Delete(name string) error {
	srv := m.services[name]
	if srv == nil {
		return nil
	}

	m.logger.Info().Str("name", name).Msg("delete service from supervisor")
	return errors.Wrapf(m.supervisor.Remove(srv.token), "failed to remove service, name=%s", name)
}

func (m *Manager) Restart(name string) (gErr error) {
	srv := m.services[name]
	if srv == nil {
		m.logger.Warn().Str("name", name).Msg("service not found, cannot restart")
		return nil
	}

	if errcheck.Check(&gErr, m.supervisor.Remove(srv.token)) {
		return errors.Wrapf(gErr, "failed to remove service, name=%s", name)
	}

	m.services[name] = &serviceWrapper{service: srv.service, token: m.supervisor.Add(srv.service)}
	m.logger.Info().Str("name", name).Msg("restarting service in supervisor")

	return nil
}

func (m *Manager) Services() []Service {
	var services []Service
	for _, srv := range m.services {
		services = append(services, srv.service)
	}
	return services
}

func (m *Manager) Serve(ctx context.Context) error {
	err := m.supervisor.Serve(ctx)

	if errors.Is(err, context.Canceled) {
		return nil
	}

	return err
}

func (m *Manager) ServeBackground(ctx context.Context) <-chan error {
	return m.supervisor.ServeBackground(ctx)
}
