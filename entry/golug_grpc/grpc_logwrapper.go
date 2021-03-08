package golug_grpc

import (
	"fmt"

	"github.com/pubgo/xlog"
	"github.com/pubgo/xlog/xlog_abc"
	"google.golang.org/grpc/grpclog"
)

func init() {
	xlog.Watch(func(logs xlog_abc.Xlog) {
		grpclog.SetLoggerV2(&loggerWrapper{log: logs.Named("grpc")})
	})
}

// loggerWrapper wraps xlog.Logger into a LoggerV2.
type loggerWrapper struct {
	log xlog.Xlog
}

// Info logs to INFO log
func (l *loggerWrapper) Info(args ...interface{}) {
	l.log.Info(sprint(args...))
}

// Infoln logs to INFO log
func (l *loggerWrapper) Infoln(args ...interface{}) {
	l.log.Info(sprint(args...))
}

// Infof logs to INFO log
func (l *loggerWrapper) Infof(format string, args ...interface{}) {
	l.log.Infof(format, args...)
}

// Warning logs to WARNING log
func (l *loggerWrapper) Warning(args ...interface{}) {
	l.log.Warn(sprint(args...))
}

// Warning logs to WARNING log
func (l *loggerWrapper) Warningln(args ...interface{}) {
	l.log.Warn(sprint(args...))
}

// Warning logs to WARNING log
func (l *loggerWrapper) Warningf(format string, args ...interface{}) {
	l.log.Warnf(format, args...)
}

// Error logs to ERROR log
func (l *loggerWrapper) Error(args ...interface{}) {
	l.log.Error(sprint(args...))
}

// Errorn logs to ERROR log
func (l *loggerWrapper) Errorln(args ...interface{}) {
	l.log.Error(sprint(args...))
}

// Errorf logs to ERROR log
func (l *loggerWrapper) Errorf(format string, args ...interface{}) {
	l.log.Errorf(format, args...)
}

// Fatal logs to ERROR log
func (l *loggerWrapper) Fatal(args ...interface{}) {
	l.log.Fatal(sprint(args...))
}

// Fatalln logs to ERROR log
func (l *loggerWrapper) Fatalln(args ...interface{}) {
	l.log.Fatal(sprint(args...))
}

// Error logs to ERROR log
func (l *loggerWrapper) Fatalf(format string, args ...interface{}) {
	l.log.Fatalf(format, args...)
}

// v returns true for all verbose level.
func (l *loggerWrapper) V(v int) bool { return true }

func sprint(args ...interface{}) string { return fmt.Sprint(args...) }
