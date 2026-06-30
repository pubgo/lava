// Package logging 提供日志工厂注册表与配置结构。
//
// 各日志后端（slog、stdlog、grpclog 等）在 init() 中通过 Register 注册 Factory，
// logbuilder 在启动时根据配置实例化对应的 Logger。
package logging

import (
	"sync"

	"github.com/pubgo/funk/v2/assert"
	"github.com/pubgo/funk/v2/log"
	"github.com/pubgo/funk/v2/recovery"
)

type Factory func(log log.Logger)

var (
	factories       = make(map[string]Factory)
	factoryMu       sync.RWMutex
	disabledLoggers = make(map[string]struct{})
	logFilePath     string // 日志文件路径，供 loggerdebug 使用
)

// List 返回所有注册的日志工厂（返回副本，线程安全）
func List() map[string]Factory {
	factoryMu.RLock()
	defer factoryMu.RUnlock()

	result := make(map[string]Factory, len(factories))
	for k, v := range factories {
		result[k] = v
	}
	return result
}

// Register 注册日志工厂
func Register(name string, factory Factory) {
	defer recovery.Exit()
	assert.If(name == "" || factory == nil, "[factory, name] should not be null")

	factoryMu.Lock()
	defer factoryMu.Unlock()

	assert.If(factories[name] != nil, "[factory] %s already exists", name)
	factories[name] = factory
}

// SetDisabledLoggers 设置要禁用的 logger 名称列表
// 这些 logger 的日志输出将被 EnableChecker 过滤掉
func SetDisabledLoggers(names []string) {
	factoryMu.Lock()
	defer factoryMu.Unlock()

	disabledLoggers = make(map[string]struct{}, len(names))
	for _, name := range names {
		disabledLoggers[name] = struct{}{}
	}
}

// IsDisabled 检查指定的 logger 是否被禁用
// 支持前缀匹配，如禁用 "grpc" 会同时禁用 "grpc", "grpc.server", "grpc.client" 等
func IsDisabled(name string) bool {
	factoryMu.RLock()
	defer factoryMu.RUnlock()

	// 精确匹配
	if _, ok := disabledLoggers[name]; ok {
		return true
	}

	// 前缀匹配
	for disabled := range disabledLoggers {
		if len(name) > len(disabled) && name[:len(disabled)] == disabled && name[len(disabled)] == '.' {
			return true
		}
	}

	return false
}

// SetLogFilePath 设置日志文件路径
func SetLogFilePath(path string) {
	factoryMu.Lock()
	defer factoryMu.Unlock()
	logFilePath = path
}

// GetLogFilePath 获取日志文件路径
func GetLogFilePath() string {
	factoryMu.RLock()
	defer factoryMu.RUnlock()
	return logFilePath
}
