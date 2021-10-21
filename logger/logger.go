package logger

import (
	"sync"

	"github.com/pubgo/dix"
	"github.com/pubgo/xerror"
	"go.uber.org/zap"

	"github.com/pubgo/lava/runenv"
)

var Discard = zap.NewNop()

func On(fn func(log *zap.Logger)) *zap.Logger {
	xerror.Exit(dix.Provider(fn))
	return zap.L()
}

var loggerMap sync.Map

// GetName 通过名字获取log
func GetName(name string) *zap.Logger {
	if val, ok := loggerMap.Load(name); ok {
		return val.(*zap.Logger)
	}

	var l = zap.L().Named(name)
	loggerMap.LoadOrStore(name, l)
	return l
}

// GetSugar 通过名字获取sugar log
func GetSugar(name string) *zap.SugaredLogger {
	if val, ok := loggerMap.Load(name); ok {
		return val.(*zap.Logger).Sugar()
	}

	var l = zap.L().Named(name)
	loggerMap.LoadOrStore(name, l)
	return l.Sugar()
}

func Init(log *zap.Logger) {
	// 业务日志
	bisLog := log.Named(runenv.Project)

	// 全局替换
	zap.ReplaceGlobals(bisLog)

	// 依赖注入
	xerror.Exit(dix.Provider(bisLog))
	xerror.Exit(dix.ProviderNs("lava", log))

	loggerMap.Range(func(key, value interface{}) bool {
		loggerMap.Store(key, bisLog.Named(key.(string)))
		return true
	})
}
