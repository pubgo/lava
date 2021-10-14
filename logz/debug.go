package logz

import (
	"github.com/pubgo/lava/logger"
	"github.com/pubgo/x/stack"
	"github.com/pubgo/xerror"
	"github.com/pubgo/xlog/xlog_config"
	"go.uber.org/zap"

	"github.com/pubgo/lava/consts"
)

var Discard = zap.NewNop()

var debugLog = func() *zap.Logger {
	cfg := xlog_config.NewDevConfig()
	cfg.EncoderConfig.EncodeCaller = "full"
	cfg.EncoderConfig.EncodeTime = consts.DefaultTimeFormat

	var log, err = cfg.Build()
	xerror.Panic(err)
	return log.Named("lava")
}()

func With(name string, args ...zap.Field) *zap.SugaredLogger {
	return debugLog.Named(name).With(args...).Sugar()
}

func WithErr(name string, err error) *zap.SugaredLogger {
	return debugLog.Named(name).With(logger.WithErr(err)...).Sugar()
}

func Named(name string, depth ...int) *zap.SugaredLogger {
	if len(depth) > 0 {
		return debugLog.Named(name).WithOptions(zap.AddCallerSkip(depth[0])).Sugar()
	}
	return debugLog.Named(name).Sugar()
}

func TryWith(name string, fn func()) *zap.SugaredLogger {
	var err error
	xerror.TryWith(&err, fn)
	if err == nil {
		return Discard.Sugar()
	}

	return Named(name).With(
		zap.String("err", err.Error()),
		zap.Any("err_stack", err),
		zap.String("fn", stack.Func(fn)),
	)
}
