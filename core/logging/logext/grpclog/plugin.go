package grpclog

import (
	"fmt"

	"github.com/pubgo/funk/log"
	"github.com/rs/zerolog"
	"google.golang.org/grpc/grpclog"

	"github.com/pubgo/lava/core/logging"
)

func grpcComponentName(args interface{}) func(e *zerolog.Event) {
	name := args.(string)
	return func(e *zerolog.Event) {
		e.Str("grpc-component", name[1:len(name)-1])
	}
}

func init() {
	logging.Register("grpcLog", SetLogger)
}

func SetLogger(logger log.Logger) {
	grpclog.SetLoggerV2(&loggerWrapper{
		log:      logger.WithName("grpc").WithCallerSkip(2),
		depthLog: logger.WithName("grpc-component").WithCallerSkip(2),
	})
}

var (
	_ grpclog.LoggerV2      = (*loggerWrapper)(nil)
	_ grpclog.DepthLoggerV2 = (*loggerWrapper)(nil)
)

type loggerWrapper struct {
	log           log.Logger
	depthLog      log.Logger
	printFilter   func(args ...interface{}) bool
	printfFilter  func(format string, args ...interface{}) bool
	printlnFilter func(args ...interface{}) bool
}

func (l *loggerWrapper) InfoDepth(depth int, args ...interface{}) {
	l.depthLog.WithCallerSkip(depth).Info().Func(grpcComponentName(args[0])).Msg(fmt.Sprint(args[1:]...))
}

func (l *loggerWrapper) WarningDepth(depth int, args ...interface{}) {
	l.depthLog.WithCallerSkip(depth).Warn().Func(grpcComponentName(args[0])).Msg(fmt.Sprint(args[1:]...))
}

func (l *loggerWrapper) ErrorDepth(depth int, args ...interface{}) {
	l.depthLog.WithCallerSkip(depth).Error().Func(grpcComponentName(args[0])).Msg(fmt.Sprint(args[1:]...))
}

func (l *loggerWrapper) FatalDepth(depth int, args ...interface{}) {
	l.depthLog.WithCallerSkip(depth).Fatal().Func(grpcComponentName(args[0])).Msg(fmt.Sprint(args[1:]...))
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

	l.log.Info().Msg(fmt.Sprint(args...))
}

func (l *loggerWrapper) Infoln(args ...interface{}) {
	if l.filterln(args) {
		return
	}

	l.log.Info().Msg(fmt.Sprint(args...))
}

func (l *loggerWrapper) Infof(format string, args ...interface{}) {
	if l.filterf(format, args...) {
		return
	}

	l.log.Info().Msg(fmt.Sprintf(format, args...))
}

func (l *loggerWrapper) Warning(args ...interface{}) {
	if l.filter(args...) {
		return
	}

	l.log.Warn().Msg(fmt.Sprint(args...))
}

func (l *loggerWrapper) Warningln(args ...interface{}) {
	if l.filterln(args) {
		return
	}

	l.log.Warn().Msg(fmt.Sprint(args...))
}

func (l *loggerWrapper) Warningf(format string, args ...interface{}) {
	if l.filterf(format, args...) {
		return
	}

	l.log.Warn().Msg(fmt.Sprintf(format, args...))
}

func (l *loggerWrapper) Error(args ...interface{}) {
	if l.filter(args...) {
		return
	}

	l.log.Error().Msg(fmt.Sprint(args...))
}

func (l *loggerWrapper) Errorln(args ...interface{}) {
	if l.filterln(args) {
		return
	}

	l.log.Error().Msg(fmt.Sprint(args...))
}

func (l *loggerWrapper) Errorf(format string, args ...interface{}) {
	if l.filterf(format, args...) {
		return
	}

	l.log.Error().Msg(fmt.Sprintf(format, args...))
}

func (l *loggerWrapper) Fatal(args ...interface{}) {
	if l.filter(args...) {
		return
	}

	l.log.Fatal().Msg(fmt.Sprint(args...))
}

func (l *loggerWrapper) Fatalln(args ...interface{}) {
	if l.filterln(args) {
		return
	}

	l.log.Fatal().Msg(fmt.Sprint(args...))
}

func (l *loggerWrapper) Fatalf(format string, args ...interface{}) {
	if l.filterf(format, args...) {
		return
	}

	l.log.Fatal().Msg(fmt.Sprintf(format, args...))
}

func (l *loggerWrapper) V(_ int) bool { return true }
