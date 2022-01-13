package logger

import (
	"github.com/pubgo/lava/types"
)

// logger 组件用于业务方的log记录, 主要用于handler中的log记录

type Logger interface {
	With(args types.M) Logger
	WithErr(err error) Logger
	Depth(depth ...int) Logger

	Debug(args ...interface{})
	Info(args ...interface{})
	Warn(args ...interface{})
	Error(args ...interface{})

	Debugf(format string, a ...interface{})
	Infof(format string, a ...interface{})
	Warnf(format string, a ...interface{})
	Errorf(format string, a ...interface{})

	Debugw(msg string, keysAndValues ...interface{})
	Infow(msg string, keysAndValues ...interface{})
	Warnw(msg string, keysAndValues ...interface{})
	Errorw(msg string, keysAndValues ...interface{})
}
