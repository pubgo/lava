package automaxprocs

import (
	"github.com/pubgo/lug/logger"
	"github.com/pubgo/xerror"
	"go.uber.org/automaxprocs/maxprocs"
	"go.uber.org/zap"
)

func init() {
	logger.On(func(logs *zap.Logger) {
		logs = logs.WithOptions(zap.AddCallerSkip(2))
		var log = maxprocs.Logger(func(s string, i ...interface{}) { logs.Sugar().Infof(s, i...) })
		xerror.ExitErr(maxprocs.Set(log)).(func())()
	})
}
