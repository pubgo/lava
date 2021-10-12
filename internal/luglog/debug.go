package luglog

import (
	"github.com/pubgo/xerror"
	"github.com/pubgo/xlog/xlog_config"
	"go.uber.org/zap"

	"github.com/pubgo/lug/consts"
)

var debugLog = func() *zap.Logger {
	cfg := xlog_config.NewDevConfig()
	cfg.EncoderConfig.EncodeCaller = "full"
	cfg.EncoderConfig.EncodeTime = consts.DefaultTimeFormat

	var log, err = cfg.Build()
	xerror.Panic(err)
	return log.Named("lug")
}()

func With(args ...interface{}) *zap.SugaredLogger { return debugLog.Sugar().With(args...) }
func Named(name string, depth ...int) *zap.SugaredLogger {
	if len(depth) > 0 {
		return debugLog.Named(name).WithOptions(zap.AddCallerSkip(depth[0])).Sugar()
	}
	return debugLog.Named(name).Sugar()
}
