package supervisor

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/pubgo/funk/v2/assert"
	"github.com/pubgo/funk/v2/log"
	"github.com/pubgo/funk/v2/recovery"
	"github.com/pubgo/funk/v2/stack"

	"github.com/pubgo/lava/v2/core/lifecycle"
	"github.com/pubgo/lava/v2/core/running"
	"github.com/pubgo/lava/v2/internal/logutil"
)

// serviceRunner 管理单个服务的运行
type serviceRunner struct {
	service Service
	config  ServiceConfig
	cancel  context.CancelFunc
	running bool // 是否正在运行
	stopped bool // 是否被手动停止
	failed  bool // 是否已失败（达到重启上限）
	done    chan struct{}

	// 重启状态跟踪
	restartCount     int           // 总重启次数
	consecFailures   int           // 连续失败次数
	windowRestarts   int           // 窗口期内重启次数
	windowStart      time.Time     // 窗口开始时间
	currentDelay     time.Duration // 当前重启延迟
	lastServiceStart time.Time     // 上次服务启动时间
}

func Default(lc lifecycle.Getter) *Manager {
	return NewManager(running.Project(), lc)
}

func NewManager(name string, lc lifecycle.Getter) *Manager {
	m := &Manager{
		name:     name,
		lc:       lc,
		services: make(map[string]*serviceRunner),
		logger:   log.GetLogger(name),
	}
	return m
}

type Manager struct {
	name     string
	lc       lifecycle.Getter
	logger   log.Logger
	mu       sync.RWMutex
	services map[string]*serviceRunner

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// GetServicesInfo returns the status info of all services
func (m *Manager) GetServicesInfo() []*ServiceInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	services := make([]*ServiceInfo, 0, len(m.services))
	for _, srv := range m.services {
		metric := srv.service.Metric()
		metric.Status = m.statusFromRunner(srv)
		services = append(services, &ServiceInfo{
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
		})
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

	metric := srv.service.Metric()
	metric.Status = m.statusFromRunner(srv)
	info := &ServiceInfo{
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
	return info, nil
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

func (m *Manager) OnClose(fn func()) {
	_ = m.Add(&onCloseService{fn: fn})
}

type onCloseService struct {
	fn func()
}

func (s *onCloseService) Name() string    { return "on-close-" + fmt.Sprintf("%p", s.fn) }
func (s *onCloseService) Error() error    { return nil }
func (s *onCloseService) String() string  { return s.Name() }
func (s *onCloseService) Metric() *Metric { return &Metric{Name: s.Name()} }
func (s *onCloseService) Serve(ctx context.Context) error {
	<-ctx.Done()
	s.fn()
	return NoRestartErr(nil)
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

	// 如果 manager 已经启动，立即启动这个服务
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

	// 停止服务
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
	m.mu.Lock()
	defer m.mu.Unlock()

	for name, srv := range m.services {
		if srv.cancel != nil {
			srv.cancel()
		}
		m.logger.Info().Str("name", name).Msg("removing service from supervisor")
	}

	m.services = make(map[string]*serviceRunner)
	return nil
}

func (m *Manager) RestartServices() error {
	m.mu.RLock()
	names := make([]string, 0, len(m.services))
	for name := range m.services {
		names = append(names, name)
	}
	m.mu.RUnlock()

	for _, name := range names {
		// restart each service individually to avoid holding lock for too long
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

	// 重置状态
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

	// 重置所有重启相关状态
	srv.failed = false
	srv.restartCount = 0
	srv.consecFailures = 0
	srv.windowRestarts = 0
	srv.windowStart = time.Now()
	srv.currentDelay = srv.config.RestartDelay

	m.logger.Info().Str("name", name).Msg("reset service restart counters")
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

// startRunner 启动一个服务 runner
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

// runService 运行服务的主循环，包含自动重启逻辑
func (m *Manager) runService(ctx context.Context, runner *serviceRunner) {
	srv := runner.service
	name := srv.Name()
	config := runner.config

	// 初始化重启状态
	m.mu.Lock()
	if runner.currentDelay == 0 {
		runner.currentDelay = config.RestartDelay
	}
	if runner.windowStart.IsZero() {
		runner.windowStart = time.Now()
	}
	m.mu.Unlock()

	for {
		// 先检查 context 是否已取消
		select {
		case <-ctx.Done():
			return
		default:
		}

		// 记录服务启动时间
		startAt := time.Now()
		m.mu.Lock()
		runner.lastServiceStart = startAt
		m.mu.Unlock()

		err := srv.Serve(ctx)

		// Serve 返回后立即检查 context
		// 如果 context 已取消，无论错误是什么都应该退出
		select {
		case <-ctx.Done():
			return
		default:
		}

		// 计算运行时长
		runDuration := time.Since(startAt)

		// 处理错误
		if err != nil {
			// context 相关错误视为正常停止，不重启
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return
			}

			// 检查是否是不重启的错误
			if IsNoRestartErr(err) || IsFatalErr(err) {
				m.mu.Lock()
				runner.failed = true
				m.mu.Unlock()
				m.logger.Error().Err(err).Str("name", name).Msg("service exited with fatal error, not restarting")
				return
			}

			// 根据重启策略决定是否重启
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
			// 正常退出
			if config.RestartPolicy == RestartOnFailure || config.RestartPolicy == RestartNever {
				m.logger.Info().Str("name", name).Msg("service exited normally, not restarting per policy")
				return
			}

			// RestartAlways 策略下正常退出也会重启
			// 如果运行了足够长的时间，重置连续失败计数
			if runDuration > config.RestartWindow {
				m.mu.Lock()
				runner.consecFailures = 0
				runner.currentDelay = config.RestartDelay
				m.mu.Unlock()
			}

			m.logger.Info().Str("name", name).Msg("service exited normally, will restart")
		}

		// 更新窗口内重启计数
		var (
			now            = time.Now()
			windowRestarts int
			restartCount   int
			delay          time.Duration
		)

		m.mu.Lock()
		if now.Sub(runner.windowStart) > config.RestartWindow {
			// 窗口已过期，重置
			runner.windowStart = now
			runner.windowRestarts = 0
			runner.currentDelay = config.RestartDelay // 重置延迟
		}
		runner.windowRestarts++
		runner.restartCount++
		windowRestarts = runner.windowRestarts
		restartCount = runner.restartCount
		delay = runner.currentDelay
		m.mu.Unlock()

		// 检查是否超过最大重启次数
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

		// 检查窗口期内重启次数
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

		// 计算并应用退避延迟
		m.logger.Info().Str("name", name).
			Dur("delay", delay).
			Int("window_restarts", windowRestarts).
			Int("total_restarts", restartCount).
			Msg("waiting before restart")

		if !waitDelay(ctx, delay) {
			return
		}

		// 应用指数退避
		m.mu.Lock()
		runner.currentDelay = time.Duration(float64(runner.currentDelay) * config.BackoffMultiplier)
		if runner.currentDelay > config.MaxRestartDelay {
			runner.currentDelay = config.MaxRestartDelay
		}
		m.mu.Unlock()
	}
}

func (m *Manager) start(ctx context.Context) error {
	defer recovery.Exit()

	// 保存 context
	m.ctx, m.cancel = context.WithCancel(ctx)

	logutil.OkOrFailed(m.logger, "start lifecycle before service", func() error {
		defer recovery.Exit()
		for _, run := range m.lc.GetBeforeStarts() {
			m.logger.Info().Msgf("running %s", stack.CallerWithFunc(run.Exec))
			assert.Exit(run.Exec(ctx))
		}
		return nil
	})

	// 启动所有服务
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

	// 取消所有服务
	if m.cancel != nil {
		m.cancel()
	}

	// 等待所有服务停止，带超时
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
	err := m.start(ctx)
	if err != nil {
		return err
	}

	<-ctx.Done()

	return m.stop(ctx)
}

func (m *Manager) Serve(ctx context.Context) error {
	err := m.start(ctx)
	if err != nil {
		return err
	}

	// 等待 context 取消
	<-ctx.Done()

	// 停止所有服务
	if m.cancel != nil {
		m.cancel()
	}
	m.wg.Wait()

	return nil
}

func (m *Manager) ServeBackground(ctx context.Context) <-chan error {
	errCh := make(chan error, 1)
	go func() {
		errCh <- m.Serve(ctx)
	}()
	return errCh
}

func normalizeServiceConfig(config ServiceConfig) ServiceConfig {
	if config.RestartPolicy < RestartAlways || config.RestartPolicy > RestartNever {
		config.RestartPolicy = RestartAlways
	}

	if config.MaxRestarts < 0 {
		config.MaxRestarts = 0
	}

	if config.MaxRestartsInWindow < 0 {
		config.MaxRestartsInWindow = 0
	}

	if config.RestartDelay <= 0 {
		config.RestartDelay = time.Second
	}

	if config.MaxRestartDelay <= 0 || config.MaxRestartDelay < config.RestartDelay {
		config.MaxRestartDelay = config.RestartDelay
	}

	if config.RestartWindow <= 0 {
		config.RestartWindow = 5 * time.Minute
	}

	if config.BackoffMultiplier < 1 {
		config.BackoffMultiplier = 1
	}

	return config
}

func waitDelay(ctx context.Context, delay time.Duration) bool {
	if delay <= 0 {
		select {
		case <-ctx.Done():
			return false
		default:
			return true
		}
	}

	timer := time.NewTimer(delay)
	defer func() {
		if !timer.Stop() {
			select {
			case <-timer.C:
			default:
			}
		}
	}()

	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}
