package logger

import (
	"sync"

	"go.uber.org/zap"

	"github.com/pubgo/lava/types"
)

type Fields []zap.Field

var globalLog = zap.L().Sugar()
var globalNext = zap.L().WithOptions(zap.AddCallerSkip(1)).Sugar()
var loggerMap sync.Map
var loggerNextMap sync.Map

// L global zap log
func L() *zap.Logger {
	return zap.L()
}

// S global zap sugared log
func S() *zap.SugaredLogger {
	return zap.S()
}

func Component(name string) *nameLogger {
	if name == "" {
		panic("[name] should not be null")
	}
	return &nameLogger{name: name}
}

var _ Logger = (*nameLogger)(nil)

type nameLogger struct{ name string }

func (t *nameLogger) WithErr(err error) Logger {
	if err == nil {
		return t
	}

	var log = getName(t.name).With(zap.String("err", err.Error()), zap.Any("err_stack", err))
	return &loggerWrapper{SugaredLogger: log}
}

func (t *nameLogger) With(args types.M) Logger {
	if args == nil || len(args) == 0 {
		return t
	}

	var fields = make([]interface{}, len(args))
	for k, v := range args {
		fields = append(fields, zap.Any(k, v))
	}

	var log = getName(t.name)
	return &loggerWrapper{SugaredLogger: log.With(fields...)}
}

func (t *nameLogger) Depth(depth ...int) Logger {
	if len(depth) == 0 || depth[0] == 0 {
		return t
	}

	var log = getName(t.name)
	return &loggerWrapper{SugaredLogger: log.Desugar().WithOptions(zap.AddCallerSkip(depth[0])).Sugar()}
}

func (t *nameLogger) Debug(args ...interface{})              { getNextLog(t.name).Debug(args...) }
func (t *nameLogger) Debugf(format string, a ...interface{}) { getNextLog(t.name).Debugf(format, a...) }
func (t *nameLogger) Debugw(msg string, keysAndValues ...interface{}) {
	getNextLog(t.name).Debugw(msg, keysAndValues...)
}
func (t *nameLogger) Infof(template string, args ...interface{}) {
	getNextLog(t.name).Infof(template, args...)
}
func (t *nameLogger) Info(args ...interface{}) { getNextLog(t.name).Info(args...) }
func (t *nameLogger) Infow(msg string, keysAndValues ...interface{}) {
	getNextLog(t.name).Infow(msg, keysAndValues...)
}

func (t *nameLogger) Errorf(template string, args ...interface{}) {
	getNextLog(t.name).Errorf(template, args...)
}

func (t *nameLogger) Error(args ...interface{}) {
	getNextLog(t.name).Error(args...)
}

func (t *nameLogger) Errorw(msg string, keysAndValues ...interface{}) {
	getNextLog(t.name).Errorw(msg, keysAndValues...)
}

func (t *nameLogger) Warnf(template string, args ...interface{}) {
	getNextLog(t.name).Warnf(template, args...)
}

func (t *nameLogger) Warn(args ...interface{}) {
	getNextLog(t.name).Warn(args...)
}

func (t *nameLogger) Warnw(msg string, keysAndValues ...interface{}) {
	getNextLog(t.name).Warnw(msg, keysAndValues...)
}

func getName(name string) *zap.SugaredLogger {
	if val, ok := loggerMap.Load(name); ok {
		return val.(*zap.SugaredLogger)
	}

	var l = globalLog.Named(name)
	loggerMap.LoadOrStore(name, l)
	return l
}

func getNextLog(name string) *zap.SugaredLogger {
	if val, ok := loggerNextMap.Load(name); ok {
		return val.(*zap.SugaredLogger)
	}

	var l = globalNext.Named(name)
	loggerNextMap.LoadOrStore(name, l)
	return l
}
