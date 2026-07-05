package supervisor

import (
	"context"
	"errors"
	"time"
)

// serviceRunner 管理单个服务的运行
type serviceRunner struct {
	service Service
	config  ServiceConfig
	cancel  context.CancelFunc
	running bool
	stopped bool
	failed  bool
	done    chan struct{}

	restartCount     int
	consecFailures   int
	windowRestarts   int
	windowStart      time.Time
	currentDelay     time.Duration
	lastServiceStart time.Time
}

func (m *Manager) startRunner(runner *serviceRunner) {
	if m.ctx == nil {
		return
	}

	ctx, cancel := context.WithCancel(m.ctx)
	done := make(chan struct{})

	runner.cancel = cancel
	runner.done = done
	runner.running = true

	m.wg.Add(1)
	go func(done chan struct{}) {
		defer m.wg.Done()
		defer close(done)
		defer func() {
			m.mu.Lock()
			if runner.done == done {
				runner.running = false
				runner.cancel = nil
				runner.done = nil
			}
			m.mu.Unlock()
		}()
		m.runService(ctx, runner)
	}(done)
}

func (m *Manager) runService(ctx context.Context, runner *serviceRunner) {
	srv := runner.service
	name := srv.Name()
	config := runner.config

	m.mu.Lock()
	if runner.currentDelay == 0 {
		runner.currentDelay = config.RestartDelay
	}
	if runner.windowStart.IsZero() {
		runner.windowStart = time.Now()
	}
	m.mu.Unlock()

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		startAt := time.Now()
		m.mu.Lock()
		runner.lastServiceStart = startAt
		m.mu.Unlock()

		err := srv.Serve(ctx)

		select {
		case <-ctx.Done():
			return
		default:
		}

		runDuration := time.Since(startAt)

		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return
			}

			if IsNoRestartErr(err) || IsFatalErr(err) {
				m.mu.Lock()
				runner.failed = true
				m.mu.Unlock()
				m.logger.Error().Err(err).Str("name", name).Msg("service exited with fatal error, not restarting")
				return
			}

			if config.RestartPolicy == RestartNever {
				m.mu.Lock()
				runner.failed = true
				m.mu.Unlock()
				m.logger.Error().Err(err).Str("name", name).Msg("service exited with error, restart policy is Never")
				return
			}

			var consecFailures int
			m.mu.Lock()
			runner.consecFailures++
			consecFailures = runner.consecFailures
			m.mu.Unlock()
			m.logger.Warn().Err(err).Str("name", name).
				Int("consec_failures", consecFailures).
				Msg("service exited with error")
		} else {
			if config.RestartPolicy == RestartOnFailure || config.RestartPolicy == RestartNever {
				m.logger.Info().Str("name", name).Msg("service exited normally, not restarting per policy")
				return
			}

			if runDuration > config.RestartWindow {
				m.mu.Lock()
				runner.consecFailures = 0
				runner.currentDelay = config.RestartDelay
				m.mu.Unlock()
			}

			m.logger.Info().Str("name", name).Msg("service exited normally, will restart")
		}

		var (
			now            = time.Now()
			windowRestarts int
			restartCount   int
			delay          time.Duration
		)

		m.mu.Lock()
		if now.Sub(runner.windowStart) > config.RestartWindow {
			runner.windowStart = now
			runner.windowRestarts = 0
			runner.currentDelay = config.RestartDelay
		}
		runner.windowRestarts++
		runner.restartCount++
		windowRestarts = runner.windowRestarts
		restartCount = runner.restartCount
		delay = runner.currentDelay
		m.mu.Unlock()

		if config.MaxRestarts > 0 && restartCount > config.MaxRestarts {
			m.mu.Lock()
			runner.failed = true
			m.mu.Unlock()
			m.logger.Error().Str("name", name).
				Int("restarts", restartCount).
				Int("max_restarts", config.MaxRestarts).
				Msg("service exceeded max restart limit, marking as failed")
			return
		}

		if config.MaxRestartsInWindow > 0 && windowRestarts > config.MaxRestartsInWindow {
			m.mu.Lock()
			runner.failed = true
			m.mu.Unlock()
			m.logger.Error().Str("name", name).
				Int("window_restarts", windowRestarts).
				Int("max_in_window", config.MaxRestartsInWindow).
				Dur("window", config.RestartWindow).
				Msg("service restart rate too high, marking as failed")
			return
		}

		m.logger.Info().Str("name", name).
			Dur("delay", delay).
			Int("window_restarts", windowRestarts).
			Int("total_restarts", restartCount).
			Msg("waiting before restart")

		if !waitDelay(ctx, delay) {
			return
		}

		m.mu.Lock()
		runner.currentDelay = time.Duration(float64(runner.currentDelay) * config.BackoffMultiplier)
		if runner.currentDelay > config.MaxRestartDelay {
			runner.currentDelay = config.MaxRestartDelay
		}
		m.mu.Unlock()
	}
}
