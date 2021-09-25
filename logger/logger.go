package logger

import (
	"github.com/pubgo/dix"
	"github.com/pubgo/xerror"

	"go.uber.org/zap"
)

func init() {
	var log = xerror.ExitErr(zap.NewDevelopment()).(*zap.Logger)
	zap.ReplaceGlobals(log)
}

func On(fn func(log *zap.Logger)) *zap.Logger {
	xerror.Exit(dix.Provider(fn))
	return zap.L()
}
