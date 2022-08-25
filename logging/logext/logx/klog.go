package klog

import (
	"github.com/go-kit/log"
	"github.com/pubgo/dix"
	"github.com/pubgo/funk/logx"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/pubgo/lava/logging"
)

func init() {
	dix.Provider(func() logging.ExtLog {
		return func(logger *logging.Logger) {
			logx.SetLog(NewZapSugarLogger(logging.ModuleLog(logger, "logx").L(), zapcore.DebugLevel))
		}
	})
}

type zapSugarLogger func(msg string, keysAndValues ...interface{})

func (l zapSugarLogger) Log(kv ...interface{}) error {
	l("", kv...)
	return nil
}

// NewZapSugarLogger returns a Go kit log.Logger that sends
// log events to a zap.Logger.
func NewZapSugarLogger(logger *zap.Logger, level zapcore.Level) log.Logger {
	sugarLogger := logger.WithOptions(zap.AddCallerSkip(2)).Sugar()
	var sugar zapSugarLogger
	switch level {
	case zapcore.DebugLevel:
		sugar = sugarLogger.Debugw
	case zapcore.InfoLevel:
		sugar = sugarLogger.Infow
	case zapcore.WarnLevel:
		sugar = sugarLogger.Warnw
	case zapcore.ErrorLevel:
		sugar = sugarLogger.Errorw
	case zapcore.DPanicLevel:
		sugar = sugarLogger.DPanicw
	case zapcore.PanicLevel:
		sugar = sugarLogger.Panicw
	case zapcore.FatalLevel:
		sugar = sugarLogger.Fatalw
	default:
		sugar = sugarLogger.Infow
	}
	return sugar
}
