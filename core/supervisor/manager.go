package supervisor

import (
	"context"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/pubgo/funk/v2/assert"
	"github.com/pubgo/funk/v2/async"
	"github.com/pubgo/funk/v2/errors"
	"github.com/pubgo/funk/v2/log"
	"github.com/pubgo/funk/v2/recovery"
	"github.com/pubgo/funk/v2/result"
	"github.com/pubgo/funk/v2/running"
	"github.com/pubgo/funk/v2/stack"
	"github.com/thejerf/suture/v4"

	"github.com/pubgo/lava/v2/core/debug"
	"github.com/pubgo/lava/v2/core/lifecycle"
	"github.com/pubgo/lava/v2/internal/logutil"
	"github.com/pubgo/lava/v2/pkg/netutil"
)

type serviceWrapper struct {
	token   suture.ServiceToken
	service Service
}

func Default(lc lifecycle.Getter) *Manager {
	return NewManager(running.Project(), lc)
}

func NewManager(name string, lc lifecycle.Getter) *Manager {
	m := &Manager{
		lc:         lc,
		supervisor: suture.New(name, SpecWithInfoLogger()),
		services:   make(map[string]*serviceWrapper),
		logger:     log.GetLogger(name),
	}
	return m.init()
}

type Manager struct {
	lc         lifecycle.Getter
	logger     log.Logger
	supervisor *Supervisor
	services   map[string]*serviceWrapper
}

func (m *Manager) init() *Manager {
	debug.Route("/supervisor", func(router fiber.Router) {
		router.Get("services", func(ctx *fiber.Ctx) error {
			services := make([]*Metric, 0, len(m.services))
			for _, srv := range m.services {
				services = append(services, srv.service.Metric())
			}
			return ctx.JSON(services)
		})
	})

	return m
}

func (m *Manager) Has(name string) bool {
	_, ok := m.services[name]
	return ok
}

func (m *Manager) OnClose(fn func()) {
	m.supervisor.Add(serviceFn(func(ctx context.Context) error {
		<-ctx.Done()
		fn()
		return nil
	}))
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
		m.logger.Warn().Str("name", name).Msg("service not found, cannot delete")
		return nil
	}

	defer func() { delete(m.services, name) }()
	m.logger.Info().Str("name", name).Msg("delete service from supervisor")
	return errors.Wrapf(m.supervisor.Remove(srv.token), "failed to remove service, name=%s", name)
}

func (m *Manager) RemoveServices() (gErr error) {
	for name, srv := range m.services {
		if result.ThrowErr(&gErr, m.supervisor.Remove(srv.token)) {
			return errors.Wrapf(gErr, "failed to remove service, name=%s", name)
		}
		m.logger.Info().Str("name", name).Msg("removing service from supervisor")
	}

	m.services = make(map[string]*serviceWrapper)
	return nil
}

func (m *Manager) RestartServices() (gErr error) {
	for name, srv := range m.services {
		if result.ThrowErr(&gErr, m.supervisor.Remove(srv.token)) {
			return errors.Wrapf(gErr, "failed to remove service, name=%s", name)
		}

		m.services[name] = &serviceWrapper{service: srv.service, token: m.supervisor.Add(srv.service)}
		m.logger.Info().Str("name", name).Msg("restarting service in supervisor")
	}

	return nil
}

func (m *Manager) RestartService(name string) (gErr error) {
	srv := m.services[name]
	if srv == nil {
		m.logger.Warn().Str("name", name).Msg("service not found, cannot restart")
		return nil
	}

	if result.ThrowErr(&gErr, m.supervisor.Remove(srv.token)) {
		return errors.Wrapf(gErr, "failed to remove service, name=%s", name)
	}

	m.services[name] = &serviceWrapper{service: srv.service, token: m.supervisor.Add(srv.service)}
	m.logger.Info().Str("name", name).Msg("restarting service in supervisor")

	return nil
}

func (m *Manager) Services() []Service {
	services := make([]Service, 0, len(m.services))
	for _, srv := range m.services {
		services = append(services, srv.service)
	}
	return services
}

func (m *Manager) start(ctx context.Context) error {
	defer recovery.Exit()
	logutil.OkOrFailed(m.logger, "start lifecycle before service", func() error {
		defer recovery.Exit()
		for _, run := range m.lc.GetBeforeStarts() {
			m.logger.Info().Msgf("running %s", stack.CallerWithFunc(run.Exec))
			assert.Exit(run.Exec(ctx))
		}
		return nil
	})

	async.GoDelay(func() error {
		err := m.supervisor.Serve(ctx)
		if netutil.IsErrServerClosed(err) {
			return nil
		}
		assert.Exit(err)
		return nil
	})

	logutil.OkOrFailed(m.logger, "start lifecycle after service", func() error {
		defer recovery.Exit()
		for _, run := range m.lc.GetAfterStarts() {
			m.logger.Info().Msgf("running %s", stack.CallerWithFunc(run.Exec))
			assert.Exit(run.Exec(ctx))
		}
		return nil
	})

	return nil
}

func (m *Manager) stop(ctx context.Context) error {
	defer recovery.DebugPrint()

	logutil.OkOrFailed(m.logger, "stop lifecycle before service", func() error {
		for _, run := range m.lc.GetBeforeStops() {
			logutil.LogOrErr(m.logger, fmt.Sprintf("running %s", stack.CallerWithFunc(run.Exec)), func() error {
				return run.Exec(ctx)
			})
		}
		return nil
	})

	unstoppedServices, _ := m.supervisor.UnstoppedServiceReport()
	if len(unstoppedServices) > 0 {
		for _, service := range unstoppedServices {
			m.logger.Error().Any("service", service).Msgf("service:%s is still running", service.Name)
		}
		return errors.New("services are still running")
	}

	logutil.OkOrFailed(m.logger, "stop lifecycle after service", func() error {
		for _, run := range m.lc.GetAfterStops() {
			logutil.LogOrErr(m.logger, fmt.Sprintf("running %s", stack.CallerWithFunc(run.Exec)), func() error {
				return run.Exec(ctx)
			})
		}
		return nil
	})

	return nil
}

func (m *Manager) Run(ctx context.Context) error {
	err := m.start(ctx)
	if err != nil {
		return err
	}

	<-ctx.Done()

	return m.stop(ctx)
}

func (m *Manager) Serve(ctx context.Context) error {
	err := m.supervisor.Serve(ctx)

	if netutil.IsErrServerClosed(err) {
		return nil
	}

	return err
}

func (m *Manager) ServeBackground(ctx context.Context) <-chan error {
	return m.supervisor.ServeBackground(ctx)
}
