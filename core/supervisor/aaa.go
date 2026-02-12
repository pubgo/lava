// Service supervision and lifecycle management for lava services
package supervisor

import (
	"context"
	"time"
)

// ServiceStatus 服务状态
type ServiceStatus string

const (
	StatusIdle     ServiceStatus = "idle"     // 空闲，未启动
	StatusRunning  ServiceStatus = "running"  // 运行中
	StatusStopped  ServiceStatus = "stopped"  // 已停止（手动）
	StatusError    ServiceStatus = "error"    // 错误状态
	StatusCrashing ServiceStatus = "crashing" // 崩溃循环中
	StatusFailed   ServiceStatus = "failed"   // 已失败（达到重启上限）
)

// Metric 服务指标
type Metric struct {
	Name             string        `json:"name"`               // 服务名称
	Status           ServiceStatus `json:"status"`             // 当前状态
	StartCount       uint32        `json:"start_count"`        // 启动次数
	ErrorCount       uint32        `json:"error_count"`        // 错误次数
	SuccessCount     uint32        `json:"success_count"`      // 成功退出次数
	ConsecFailures   uint32        `json:"consec_failures"`    // 连续失败次数
	LastError        string        `json:"last_error"`         // 最后一次错误信息
	LastErrorTime    time.Time     `json:"last_error_time"`    // 最后一次错误时间
	LastStartTime    time.Time     `json:"last_start_time"`    // 最后一次启动时间
	LastStopTime     time.Time     `json:"last_stop_time"`     // 最后一次停止时间
	CurrentUptime    time.Duration `json:"current_uptime"`     // 当前运行时长
	TotalUptime      time.Duration `json:"total_uptime"`       // 总运行时长
	AverageUptime    time.Duration `json:"average_uptime"`     // 平均运行时长
	CreatedAt        time.Time     `json:"created_at"`         // 服务创建时间
	CurrentDelay     time.Duration `json:"current_delay"`      // 当前重启延迟
	RestartsInWindow uint32        `json:"restarts_in_window"` // 窗口期内重启次数
}

// Service 服务接口
type Service interface {
	Name() string
	Error() error
	String() string
	Serve(ctx context.Context) error
	Metric() *Metric
}

// serviceFn 函数类型服务
type serviceFn func(ctx context.Context) error

func (fn serviceFn) Serve(ctx context.Context) error {
	return fn(ctx)
}

// RestartPolicy 重启策略
type RestartPolicy int

const (
	// RestartAlways 总是重启（除非手动停止）
	RestartAlways RestartPolicy = iota
	// RestartOnFailure 仅在失败时重启
	RestartOnFailure
	// RestartNever 从不自动重启
	RestartNever
)

// ServiceConfig 服务配置
type ServiceConfig struct {
	// RestartPolicy 重启策略
	RestartPolicy RestartPolicy
	// MaxRestarts 最大重启次数，0 表示无限制
	MaxRestarts int
	// RestartDelay 初始重启延迟
	RestartDelay time.Duration
	// MaxRestartDelay 最大重启延迟（用于指数退避）
	MaxRestartDelay time.Duration
	// RestartWindow 重启计数窗口期，在此期间内的重启会被计数
	// 如果服务运行超过此时间后崩溃，重启计数会重置
	RestartWindow time.Duration
	// MaxRestartsInWindow 窗口期内最大重启次数，超过则标记为 failed
	MaxRestartsInWindow int
	// BackoffMultiplier 退避乘数，默认 2.0
	BackoffMultiplier float64
	// AutoStart 是否自动启动，默认 true
	AutoStart bool
}

// DefaultServiceConfig 默认服务配置
func DefaultServiceConfig() ServiceConfig {
	return ServiceConfig{
		AutoStart:           true,
		RestartPolicy:       RestartAlways,
		MaxRestarts:         0,               // 无限制
		RestartDelay:        time.Second,     // 初始 1 秒
		MaxRestartDelay:     time.Minute,     // 最大 1 分钟
		RestartWindow:       5 * time.Minute, // 5 分钟窗口
		MaxRestartsInWindow: 5,               // 窗口内最多 5 次
		BackoffMultiplier:   2.0,             // 每次翻倍
	}
}

// Option 配置选项
type Option func(*ServiceConfig)

// WithAutoStart 设置是否自动启动
func WithAutoStart(autoStart bool) Option {
	return func(c *ServiceConfig) {
		c.AutoStart = autoStart
	}
}

// WithRestartPolicy 设置重启策略
func WithRestartPolicy(policy RestartPolicy) Option {
	return func(c *ServiceConfig) {
		c.RestartPolicy = policy
	}
}

// WithMaxRestarts 设置最大重启次数
func WithMaxRestarts(n int) Option {
	return func(c *ServiceConfig) {
		c.MaxRestarts = n
	}
}

// WithRestartDelay 设置重启延迟
func WithRestartDelay(d time.Duration) Option {
	return func(c *ServiceConfig) {
		c.RestartDelay = d
	}
}

// WithBackoff 设置退避策略
func WithBackoff(maxDelay time.Duration, multiplier float64) Option {
	return func(c *ServiceConfig) {
		c.MaxRestartDelay = maxDelay
		c.BackoffMultiplier = multiplier
	}
}

// WithRestartWindow 设置重启窗口
func WithRestartWindow(window time.Duration, maxRestarts int) Option {
	return func(c *ServiceConfig) {
		c.RestartWindow = window
		c.MaxRestartsInWindow = maxRestarts
	}
}

// ServiceInfo 包含服务指标和运行时状态
type ServiceInfo struct {
	*Metric
	Stopped          bool          `json:"stopped"`
	Failed           bool          `json:"failed"`
	RestartCount     int           `json:"restart_count"`
	ConsecFailures   int           `json:"consec_failures"`
	WindowRestarts   int           `json:"window_restarts"`
	CurrentDelay     time.Duration `json:"current_delay_ns"`
	CurrentDelayStr  string        `json:"current_delay"`
	WindowStart      time.Time     `json:"window_start"`
	LastServiceStart time.Time     `json:"last_service_start"`
}
