package supervisor

import (
	"context"
	"fmt"
	"time"

	"github.com/pubgo/funk/v2/assert"
	"github.com/pubgo/funk/v2/recovery"
	"github.com/pubgo/funk/v2/stack"

	"github.com/pubgo/lava/v2/internal/logutil"
)

func (m *Manager) start(ctx context.Context) error {
	defer recovery.Exit()

	m.ctx, m.cancel = context.WithCancel(ctx)

	logutil.OkOrFailed(m.logger, "start lifecycle before service", func() error {
		defer recovery.Exit()
		for _, run := range m.lc.GetBeforeStarts() {
			m.logger.Info().Msgf("running %s", stack.CallerWithFunc(run.Exec))
			assert.Exit(run.Exec(ctx))
		}
		return nil
	})

	m.mu.Lock()
	for _, runner := range m.services {
		if !runner.stopped {
			m.startRunner(runner)
		}
	}
	m.mu.Unlock()

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

	if m.cancel != nil {
		m.cancel()
	}

	done := make(chan struct{})
	go func() {
		m.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		m.logger.Info().Msg("all services stopped")
	case <-time.After(30 * time.Second):
		m.logger.Warn().Msg("timeout waiting for services to stop")
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
	if err := m.start(ctx); err != nil {
		return err
	}

	<-ctx.Done()

	return m.stop(ctx)
}

func (m *Manager) Serve(ctx context.Context) error {
	if err := m.start(ctx); err != nil {
		return err
	}

	<-ctx.Done()

	return m.stop(ctx)
}

func (m *Manager) ServeBackground(ctx context.Context) <-chan error {
	errCh := make(chan error, 1)
	go func() {
		errCh <- m.Serve(ctx)
	}()
	return errCh
}
