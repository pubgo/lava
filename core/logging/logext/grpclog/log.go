package grpclog

import (
	"fmt"

	"github.com/pubgo/funk/v2/log"
	"github.com/rs/zerolog"
	"google.golang.org/grpc/grpclog"

	"github.com/pubgo/lava/v2/core/logging"
)

const (
	grpcLvlInfo int = iota
	grpcLvlWarn
	grpcLvlError
	grpcLvlFatal
)

// _grpcToZapLevel maps gRPC log levels to zap log levels.
var _grpcToZapLevel = map[int]zerolog.Level{
	grpcLvlInfo:  zerolog.InfoLevel,
	grpcLvlWarn:  zerolog.WarnLevel,
	grpcLvlError: zerolog.ErrorLevel,
	grpcLvlFatal: zerolog.FatalLevel,
}

func grpcComponentName(args any) func(e *zerolog.Event) {
	name := args.(string)
	return func(e *zerolog.Event) {
		e.Str("grpc-component", name[1:len(name)-1])
	}
}

func init() {
	logging.Register("grpcLog", SetLogger)
}

func SetLogger(logger log.Logger) {
	logger = logger.WithName("grpc")
	grpclog.SetLoggerV2(&loggerWrapper{
		log:      logger.WithCallerSkip(2),
		depthLog: logger,
	})
}

var (
	_ grpclog.LoggerV2      = (*loggerWrapper)(nil)
	_ grpclog.DepthLoggerV2 = (*loggerWrapper)(nil)
)

type loggerWrapper struct {
	log           log.Logger
	depthLog      log.Logger
	printFilter   func(args ...any) bool
	printfFilter  func(format string, args ...any) bool
	printlnFilter func(args ...any) bool
}

// DepthLoggerV2 实现

func (l *loggerWrapper) InfoDepth(depth int, args ...any) {
	l.depthLog.WithCallerSkip(depth + 2).Info().Func(grpcComponentName(args[0])).Msg(fmt.Sprint(args[1:]...))
}

func (l *loggerWrapper) WarningDepth(depth int, args ...any) {
	l.depthLog.WithCallerSkip(depth + 2).Warn().Func(grpcComponentName(args[0])).Msg(fmt.Sprint(args[1:]...))
}

func (l *loggerWrapper) ErrorDepth(depth int, args ...any) {
	l.depthLog.WithCallerSkip(depth + 2).Error().Func(grpcComponentName(args[0])).Msg(fmt.Sprint(args[1:]...))
}

func (l *loggerWrapper) FatalDepth(depth int, args ...any) {
	l.depthLog.WithCallerSkip(depth + 2).Fatal().Func(grpcComponentName(args[0])).Msg(fmt.Sprint(args[1:]...))
}

// Filter 设置

func (l *loggerWrapper) SetPrintFilter(filter func(args ...any) bool) {
	l.printFilter = filter
}

func (l *loggerWrapper) SetPrintfFilter(filter func(format string, args ...any) bool) {
	l.printfFilter = filter
}

func (l *loggerWrapper) SetPrintlnFilter(filter func(args ...any) bool) {
	l.printlnFilter = filter
}

func (l *loggerWrapper) filter(args ...any) bool {
	return l.printFilter != nil && l.printFilter(args...)
}

func (l *loggerWrapper) filterf(format string, args ...any) bool {
	return l.printfFilter != nil && l.printfFilter(format, args...)
}

func (l *loggerWrapper) filterln(args ...any) bool {
	return l.printlnFilter != nil && l.printlnFilter(args...)
}

// LoggerV2 实现 - Info

func (l *loggerWrapper) Info(args ...any) {
	if l.filter(args) {
		return
	}
	l.log.Info().Msg(fmt.Sprint(args...))
}

func (l *loggerWrapper) Infoln(args ...any) {
	if l.filterln(args) {
		return
	}
	l.log.Info().Msg(fmt.Sprint(args...))
}

func (l *loggerWrapper) Infof(format string, args ...any) {
	if l.filterf(format, args...) {
		return
	}
	l.log.Info().Msgf(format, args...)
}

// LoggerV2 实现 - Warning

func (l *loggerWrapper) Warning(args ...any) {
	if l.filter(args...) {
		return
	}
	l.log.Warn().Msg(fmt.Sprint(args...))
}

func (l *loggerWrapper) Warningln(args ...any) {
	if l.filterln(args) {
		return
	}
	l.log.Warn().Msg(fmt.Sprint(args...))
}

func (l *loggerWrapper) Warningf(format string, args ...any) {
	if l.filterf(format, args...) {
		return
	}
	l.log.Warn().Msgf(format, args...)
}

// LoggerV2 实现 - Error

func (l *loggerWrapper) Error(args ...any) {
	if l.filter(args...) {
		return
	}
	l.log.Error().Msg(fmt.Sprint(args...))
}

func (l *loggerWrapper) Errorln(args ...any) {
	if l.filterln(args) {
		return
	}
	l.log.Error().Msg(fmt.Sprint(args...))
}

func (l *loggerWrapper) Errorf(format string, args ...any) {
	if l.filterf(format, args...) {
		return
	}
	l.log.Error().Msgf(format, args...)
}

// LoggerV2 实现 - Fatal

func (l *loggerWrapper) Fatal(args ...any) {
	if l.filter(args...) {
		return
	}
	l.log.Fatal().Msg(fmt.Sprint(args...))
}

func (l *loggerWrapper) Fatalln(args ...any) {
	if l.filterln(args) {
		return
	}
	l.log.Fatal().Msg(fmt.Sprint(args...))
}

func (l *loggerWrapper) Fatalf(format string, args ...any) {
	if l.filterf(format, args...) {
		return
	}
	l.log.Fatal().Msgf(format, args...)
}

// V 实现日志级别检查

func (l *loggerWrapper) V(level int) bool {
	return _grpcToZapLevel[level] >= zerolog.GlobalLevel()
}
