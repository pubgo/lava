package db

import (
	"fmt"

	"go.uber.org/zap"
	xormLog "xorm.io/xorm/log"
)

// newLogger init a log bridge for xorm
func newLogger() xormLog.Logger {
	return &logBridge{
		logger: zap.L().Named(Name + "_tracing").WithOptions(zap.AddCallerSkip(1)),
	}
}

var _ xormLog.ContextLogger = (*logBridge)(nil)

// logBridge a logger bridge from Logger to xorm
type logBridge struct {
	lvl     xormLog.LogLevel
	showSQL bool
	logger  *zap.Logger
}

func (l *logBridge) BeforeSQL(ctx xormLog.LogContext) {}
func (l *logBridge) AfterSQL(ctx xormLog.LogContext) {
	var sessionPart string
	v := ctx.Ctx.Value(xormLog.SessionIDKey)
	if key, ok := v.(string); ok {
		sessionPart = fmt.Sprintf(" [%s]", key)
	}

	if ctx.ExecuteTime > 0 {
		l.logger.Sugar().Infof("[SQL]%s [%s %v] - %v", sessionPart, ctx.SQL, ctx.Args, ctx.ExecuteTime)
	} else {
		l.logger.Sugar().Infof("[SQL]%s [%s %v]", sessionPart, ctx.SQL, ctx.Args)
	}
}

// Debug show debug log
func (l *logBridge) Debug(v ...interface{}) {
	l.logger.Debug(fmt.Sprint(v...))
}

// Debugf show debug log
func (l *logBridge) Debugf(format string, v ...interface{}) {
	l.logger.Sugar().Debugf(format, v...)
}

// Error show error log
func (l *logBridge) Error(v ...interface{}) {
	l.logger.Error(fmt.Sprint(v...))
}

// Errorf show error log
func (l *logBridge) Errorf(format string, v ...interface{}) {
	l.logger.Sugar().Errorf(format, v...)
}

// Info show information level log
func (l *logBridge) Info(v ...interface{}) {
	l.logger.Info(fmt.Sprint(v...))
}

// Infof show information level log
func (l *logBridge) Infof(format string, v ...interface{}) {
	l.logger.Sugar().Infof(format, v...)
}

// Warn show warning log
func (l *logBridge) Warn(v ...interface{}) {
	l.logger.Warn(fmt.Sprint(v...))
}

// Warnf show warnning log
func (l *logBridge) Warnf(format string, v ...interface{}) {
	l.logger.Sugar().Warnf(format, v...)
}

// Level get logger level
func (l *logBridge) Level() xormLog.LogLevel { return l.lvl }

// SetLevel set the logger level
func (l *logBridge) SetLevel(lvl xormLog.LogLevel) { l.lvl = lvl }

// ShowSQL set if record SQL
func (l *logBridge) ShowSQL(show ...bool) {
	if len(show) > 0 {
		l.showSQL = show[0]
	} else {
		l.showSQL = true
	}
}

// IsShowSQL if record SQL
func (l *logBridge) IsShowSQL() bool { return l.showSQL }
