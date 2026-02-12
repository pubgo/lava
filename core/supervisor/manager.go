package supervisor

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/pubgo/funk/v2/assert"
	"github.com/pubgo/funk/v2/log"
	"github.com/pubgo/funk/v2/recovery"
	"github.com/pubgo/funk/v2/running"
	"github.com/pubgo/funk/v2/stack"

	"github.com/pubgo/lava/v2/core/lifecycle"
	"github.com/pubgo/lava/v2/internal/logutil"
)

// serviceRunner 管理单个服务的运行
type serviceRunner struct {
	service Service
	config  ServiceConfig
	cancel  context.CancelFunc
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
		// 覆盖状态
		if srv.failed {
			metric.Status = StatusFailed
		} else if srv.stopped {
			metric.Status = StatusStopped
		} else if srv.consecFailures > 0 {
			metric.Status = StatusCrashing
		}
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
		return nil, fmt.Errorf("service not found: %s", name)
	}

	metric := srv.service.Metric()
	if srv.failed {
		metric.Status = StatusFailed
	} else if srv.stopped {
		metric.Status = StatusStopped
	} else if srv.consecFailures > 0 {
		metric.Status = StatusCrashing
	}
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
	name := srv.Name()
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.services[name]; ok {
		return fmt.Errorf("service already exists, name=%s", name)
	}

	m.logger.Info().Str("name", name).Msg("add service to supervisor")

	config := DefaultServiceConfig()
	for _, opt := range opts {
		opt(&config)
	}

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

// AddWithConfig 添加服务并指定配置
func (m *Manager) AddWithConfig(srv Service, config ServiceConfig) error {
	name := srv.Name()
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.services[name]; ok {
		return fmt.Errorf("service already exists, name=%s", name)
	}

	m.logger.Info().Str("name", name).Msg("add service to supervisor with config")
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
		return fmt.Errorf("service not found, name=%s", name)
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
	m.mu.Lock()
	defer m.mu.Unlock()

	for name, srv := range m.services {
		m.restartRunnerLocked(srv)
		m.logger.Info().Str("name", name).Msg("restarting service in supervisor")
	}

	return nil
}

func (m *Manager) RestartService(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	srv := m.services[name]
	if srv == nil {
		m.logger.Warn().Str("name", name).Msg("service not found, cannot restart")
		return fmt.Errorf("service not found, name=%s", name)
	}

	// 如果服务已暂停，直接启动
	if srv.stopped {
		srv.stopped = false
		m.startRunner(srv)
		m.logger.Info().Str("name", name).Msg("starting stopped service in supervisor")
		return nil
	}

	m.restartRunnerLocked(srv)
	m.logger.Info().Str("name", name).Msg("restarting service in supervisor")

	return nil
}

// restartRunnerLocked 重启服务，必须在持有锁的情况下调用
func (m *Manager) restartRunnerLocked(runner *serviceRunner) {
	// 停止旧的
	if runner.cancel != nil {
		runner.cancel()
		// 等待旧的 goroutine 退出
		if runner.done != nil {
			<-runner.done
		}
	}
	runner.stopped = false
	// 启动新的
	m.startRunner(runner)
}

// StopService 暂停服务，但保留在 services map 中
func (m *Manager) StopService(name string) error {
	m.mu.Lock()
	srv := m.services[name]
	if srv == nil {
		m.mu.Unlock()
		m.logger.Warn().Str("name", name).Msg("service not found, cannot stop")
		return fmt.Errorf("service not found, name=%s", name)
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
		return fmt.Errorf("service not found, name=%s", name)
	}

	if !srv.stopped && !srv.failed {
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
		return fmt.Errorf("service not found, name=%s", name)
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

// GetServiceStatus 获取服务运行状态
func (m *Manager) GetServiceStatus(name string) (status ServiceStatus, info map[string]any, err error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	srv := m.services[name]
	if srv == nil {
		return "", nil, fmt.Errorf("service not found, name=%s", name)
	}

	info = map[string]any{
		"stopped":            srv.stopped,
		"failed":             srv.failed,
		"restart_count":      srv.restartCount,
		"consec_failures":    srv.consecFailures,
		"window_restarts":    srv.windowRestarts,
		"current_delay":      srv.currentDelay.String(),
		"window_start":       srv.windowStart,
		"last_service_start": srv.lastServiceStart,
	}

	if srv.failed {
		return StatusFailed, info, nil
	}
	if srv.stopped {
		return StatusStopped, info, nil
	}
	if srv.consecFailures > 0 {
		return StatusCrashing, info, nil
	}
	if srv.cancel != nil {
		return StatusRunning, info, nil
	}
	return StatusIdle, info, nil
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
	runner.cancel = cancel
	runner.done = make(chan struct{})

	m.wg.Add(1)
	go func() {
		defer m.wg.Done()
		defer close(runner.done)
		m.runService(ctx, runner)
	}()
}

// runService 运行服务的主循环，包含自动重启逻辑
func (m *Manager) runService(ctx context.Context, runner *serviceRunner) {
	srv := runner.service
	name := srv.Name()
	config := runner.config

	// 初始化重启状态
	if runner.currentDelay == 0 {
		runner.currentDelay = config.RestartDelay
	}
	if runner.windowStart.IsZero() {
		runner.windowStart = time.Now()
	}

	for {
		// 先检查 context 是否已取消
		select {
		case <-ctx.Done():
			return
		default:
		}

		// 记录服务启动时间
		runner.lastServiceStart = time.Now()

		err := srv.Serve(ctx)

		// Serve 返回后立即检查 context
		// 如果 context 已取消，无论错误是什么都应该退出
		select {
		case <-ctx.Done():
			return
		default:
		}

		// 计算运行时长
		runDuration := time.Since(runner.lastServiceStart)

		// 处理错误
		if err != nil {
			// context 相关错误视为正常停止，不重启
			if err == context.Canceled || err == context.DeadlineExceeded {
				return
			}

			// 检查是否是不重启的错误
			if IsNoRestartErr(err) || IsFatalErr(err) {
				runner.failed = true
				m.logger.Error().Err(err).Str("name", name).Msg("service exited with fatal error, not restarting")
				return
			}

			// 根据重启策略决定是否重启
			if config.RestartPolicy == RestartNever {
				runner.failed = true
				m.logger.Error().Err(err).Str("name", name).Msg("service exited with error, restart policy is Never")
				return
			}

			runner.consecFailures++
			m.logger.Warn().Err(err).Str("name", name).
				Int("consec_failures", runner.consecFailures).
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
				runner.consecFailures = 0
				runner.currentDelay = config.RestartDelay
			}

			m.logger.Info().Str("name", name).Msg("service exited normally, will restart")
		}

		// 更新窗口内重启计数
		now := time.Now()
		if now.Sub(runner.windowStart) > config.RestartWindow {
			// 窗口已过期，重置
			runner.windowStart = now
			runner.windowRestarts = 0
			runner.currentDelay = config.RestartDelay // 重置延迟
		}
		runner.windowRestarts++
		runner.restartCount++

		// 检查是否超过最大重启次数
		if config.MaxRestarts > 0 && runner.restartCount >= config.MaxRestarts {
			runner.failed = true
			m.logger.Error().Str("name", name).
				Int("restarts", runner.restartCount).
				Int("max_restarts", config.MaxRestarts).
				Msg("service exceeded max restart limit, marking as failed")
			return
		}

		// 检查窗口期内重启次数
		if config.MaxRestartsInWindow > 0 && runner.windowRestarts > config.MaxRestartsInWindow {
			runner.failed = true
			m.logger.Error().Str("name", name).
				Int("window_restarts", runner.windowRestarts).
				Int("max_in_window", config.MaxRestartsInWindow).
				Dur("window", config.RestartWindow).
				Msg("service restart rate too high, marking as failed")
			return
		}

		// 计算并应用退避延迟
		delay := runner.currentDelay
		m.logger.Info().Str("name", name).
			Dur("delay", delay).
			Int("window_restarts", runner.windowRestarts).
			Int("total_restarts", runner.restartCount).
			Msg("waiting before restart")

		select {
		case <-ctx.Done():
			return
		case <-time.After(delay):
			// 应用指数退避
			runner.currentDelay = time.Duration(float64(runner.currentDelay) * config.BackoffMultiplier)
			if runner.currentDelay > config.MaxRestartDelay {
				runner.currentDelay = config.MaxRestartDelay
			}
		}
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
	m.mu.RLock()
	for _, runner := range m.services {
		if !runner.stopped {
			m.startRunner(runner)
		}
	}
	m.mu.RUnlock()

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
