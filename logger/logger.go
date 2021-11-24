package logger

import (
	"sync"

	"go.uber.org/zap"
)

type Fields []zap.Field

var Discard = zap.NewNop()
var globalLog = zap.L()
var loggerMap sync.Map

func New(name string) *nameLogger {
	if name == "" {
		panic("[name] should not be null")
	}
	return &nameLogger{name: name}
}

type nameLogger struct {
	name string
}

func (t *nameLogger) With(args ...zap.Field) *zap.Logger {
	return getName(t.name).With(args...)
}

func (t *nameLogger) Depth(depth ...int) *zap.Logger {
	if len(depth) > 0 {
		return getName(t.name).WithOptions(zap.AddCallerSkip(depth[0]))
	}
	return getName(t.name)
}

func (t *nameLogger) DepthS(depth ...int) *zap.SugaredLogger {
	return t.Depth(depth...).Sugar()
}

func (t *nameLogger) Infof(template string, args ...interface{}) {
	t.DepthS(1).Infof(template, args...)
}

func (t *nameLogger) Info(args ...interface{}) {
	t.DepthS(1).Info(args...)
}

func (t *nameLogger) Infow(msg string, keysAndValues ...interface{}) {
	t.DepthS(1).Infow(msg, keysAndValues...)
}

func (t *nameLogger) Errorf(template string, args ...interface{}) {
	t.DepthS(1).Errorf(template, args...)
}

func (t *nameLogger) Error(args ...interface{}) {
	t.DepthS(1).Error(args...)
}

func (t *nameLogger) Errorw(msg string, keysAndValues ...interface{}) {
	t.DepthS(1).Errorw(msg, keysAndValues...)
}

func (t *nameLogger) Warnf(template string, args ...interface{}) {
	t.DepthS(1).Warnf(template, args...)
}

func (t *nameLogger) Warn(args ...interface{}) {
	t.DepthS(1).Warn(args...)
}

func (t *nameLogger) Warnw(msg string, keysAndValues ...interface{}) {
	t.DepthS(1).Warnw(msg, keysAndValues...)
}

func getName(name string) *zap.Logger {
	if val, ok := loggerMap.Load(name); ok {
		return val.(*zap.Logger)
	}

	var l = globalLog.Named(name)
	loggerMap.LoadOrStore(name, l)
	return l
}
