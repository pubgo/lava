package grpclog

import (
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc/grpclog"

	"github.com/pubgo/lava/logz"
)

func init() {
	logz.On(func(z *logz.Log) {
		grpclog.SetLoggerV2(&loggerWrapper{
			log:      logz.Component("grpc").Depth(4),
			depthLog: logz.Component("grpc-component").Depth(2),
		})
	})
}

var _ grpclog.LoggerV2 = (*loggerWrapper)(nil)
var _ grpclog.DepthLoggerV2 = (*loggerWrapper)(nil)

type loggerWrapper struct {
	log           *zap.Logger
	depthLog      *zap.Logger
	printFilter   func(args ...interface{}) bool
	printfFilter  func(format string, args ...interface{}) bool
	printlnFilter func(args ...interface{}) bool
}

func (l *loggerWrapper) InfoDepth(depth int, args ...interface{}) {
	l.depthLog.WithOptions(zap.AddCallerSkip(depth)).Info(fmt.Sprint(args...))
}

func (l *loggerWrapper) WarningDepth(depth int, args ...interface{}) {
	l.depthLog.WithOptions(zap.AddCallerSkip(depth)).Warn(fmt.Sprint(args...))
}

func (l *loggerWrapper) ErrorDepth(depth int, args ...interface{}) {
	l.depthLog.WithOptions(zap.AddCallerSkip(depth)).Error(fmt.Sprint(args...))
}

func (l *loggerWrapper) FatalDepth(depth int, args ...interface{}) {
	l.depthLog.WithOptions(zap.AddCallerSkip(depth)).Fatal(fmt.Sprint(args...))
}

func (l *loggerWrapper) SetPrintFilter(filter func(args ...interface{}) bool) {
	l.printFilter = filter
}
func (l *loggerWrapper) SetPrintfFilter(filter func(format string, args ...interface{}) bool) {
	l.printfFilter = filter
}
func (l *loggerWrapper) SetPrintlnFilter(filter func(args ...interface{}) bool) {
	l.printlnFilter = filter
}

func (l *loggerWrapper) filter(args ...interface{}) bool {
	return l.printFilter != nil && l.printFilter(args...)
}

func (l *loggerWrapper) filterf(format string, args ...interface{}) bool {
	return l.printfFilter != nil && l.printfFilter(format, args...)
}

func (l *loggerWrapper) filterln(args ...interface{}) bool {
	return l.printlnFilter != nil && l.printlnFilter(args...)
}

func (l *loggerWrapper) Info(args ...interface{}) {
	if l.filter(args) {
		return
	}

	l.log.Info(fmt.Sprint(args...))
}

func (l *loggerWrapper) Infoln(args ...interface{}) {
	if l.filterln(args) {
		return
	}
	l.log.Info(fmt.Sprintln(args...))
}

func (l *loggerWrapper) Infof(format string, args ...interface{}) {
	if l.filterf(format, args...) {
		return
	}
	l.log.Info(fmt.Sprintf(format, args...))
}

func (l *loggerWrapper) Warning(args ...interface{}) {
	if l.filter(args...) {
		return
	}

	l.log.Warn(fmt.Sprint(args...))
}

func (l *loggerWrapper) Warningln(args ...interface{}) {
	if l.filterln(args) {
		return
	}

	l.log.Warn(fmt.Sprintln(args...))
}

func (l *loggerWrapper) Warningf(format string, args ...interface{}) {
	if l.filterf(format, args...) {
		return
	}

	l.log.Warn(fmt.Sprintf(format, args...))
}

func (l *loggerWrapper) Error(args ...interface{}) {
	if l.filter(args...) {
		return
	}

	l.log.Error(fmt.Sprint(args...))
}

func (l *loggerWrapper) Errorln(args ...interface{}) {
	if l.filterln(args) {
		return
	}

	l.log.Error(fmt.Sprintln(args...))
}

func (l *loggerWrapper) Errorf(format string, args ...interface{}) {
	if l.filterf(format, args...) {
		return
	}
	l.log.Error(fmt.Sprintf(format, args...))
}

func (l *loggerWrapper) Fatal(args ...interface{}) {
	if l.filter(args...) {
		return
	}

	l.log.Fatal(fmt.Sprint(args...))
}

func (l *loggerWrapper) Fatalln(args ...interface{}) {
	if l.filterln(args) {
		return
	}

	l.log.Fatal(fmt.Sprintln(args...))
}

func (l *loggerWrapper) Fatalf(format string, args ...interface{}) {
	if l.filterf(format, args...) {
		return
	}

	l.log.Fatal(fmt.Sprintf(format, args...))
}

func (l *loggerWrapper) V(_ int) bool { return true }
func (l *loggerWrapper) Lvl(lvl int) grpclog.LoggerV2 {
	return &loggerWrapper{log: l.log.WithOptions(zap.IncreaseLevel(zapcore.Level(lvl)))}
}
