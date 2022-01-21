package logging

import (
	"github.com/pubgo/dix"
	"github.com/pubgo/xerror"
	"go.uber.org/zap"
)

type Fields = []zap.Field

// L global zap log
func L() *zap.Logger {
	return zap.L()
}

// S global zap sugared log
func S() *zap.SugaredLogger {
	return zap.S()
}

type Event struct{}

// On log 依赖注入
func On(fn func(*Event)) {
	xerror.Exit(dix.Provider(fn))
}
