package logger

import (
	"sync"

	"go.uber.org/zap"
)

type Fields []zap.Field

var Discard = zap.NewNop()
var globalLog = zap.L()
var loggerMap sync.Map

func New(name string) *Logger {
	if name == "" {
		panic("[name] should not be null")
	}
	return &Logger{name: name}
}

type Logger struct {
	name string
}

func (t *Logger) With(args ...zap.Field) *zap.Logger {
	return getName(t.name).With(args...)
}

func (t *Logger) Depth(depth ...int) *zap.Logger {
	if len(depth) > 0 {
		return getName(t.name).WithOptions(zap.AddCallerSkip(depth[0]))
	}
	return getName(t.name)
}

func (t *Logger) DepthS(depth ...int) *zap.SugaredLogger {
	return t.Depth(depth...).Sugar()
}

func (t *Logger) Infof(template string, args ...interface{}) {
	t.DepthS(1).Infof(template, args...)
}

func (t *Logger) Info(args ...interface{}) {
	t.DepthS(1).Info(args...)
}

func (t *Logger) Infow(msg string, keysAndValues ...interface{}) {
	t.DepthS(1).Infow(msg, keysAndValues...)
}

func (t *Logger) Errorf(template string, args ...interface{}) {
	t.DepthS(1).Errorf(template, args...)
}

func (t *Logger) Error(args ...interface{}) {
	t.DepthS(1).Error(args...)
}

func (t *Logger) Errorw(msg string, keysAndValues ...interface{}) {
	t.DepthS(1).Errorw(msg, keysAndValues...)
}

func (t *Logger) Warnf(template string, args ...interface{}) {
	t.DepthS(1).Warnf(template, args...)
}

func (t *Logger) Warn(args ...interface{}) {
	t.DepthS(1).Warn(args...)
}

func (t *Logger) Warnw(msg string, keysAndValues ...interface{}) {
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
