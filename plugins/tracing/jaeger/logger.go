package jaeger

import (
	jLog "github.com/uber/jaeger-client-go/log"
	"go.uber.org/zap"
)

var _ jLog.Logger = (*logger)(nil)

func newLog(name string) *logger {
	return &logger{logs: zap.L().Named(name).WithOptions(zap.AddCallerSkip(4), zap.Fields(zap.Bool("tracing", true)))}
}

type logger struct {
	logs *zap.Logger
}

func (l logger) Write(p []byte) (n int, err error) {
	l.logs.Info(string(p))
	return 0, err
}

func (l logger) Debugf(msg string, args ...interface{}) {
	l.logs.Sugar().Debugf(msg, args...)
}

func (l logger) Error(msg string) {
	l.logs.Error(msg)
}

func (l logger) Infof(msg string, args ...interface{}) {
	l.logs.Sugar().Infof(msg, args...)
}
