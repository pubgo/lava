package jaeger

import (
	"github.com/pubgo/xlog"
	jLog "github.com/uber/jaeger-client-go/log"
	"go.uber.org/zap"
)

var _ jLog.Logger = (*logger)(nil)

func newLog(name string) *logger {
	return &logger{logs: xlog.GetLogger(name,
		zap.AddCallerSkip(2),
		zap.Fields(zap.Bool("tracing", true)))}
}

type logger struct {
	logs xlog.Xlog
}

func (l logger) Write(p []byte) (n int, err error) {
	l.logs.Info(string(p))
	return 0, err
}

func (l logger) Debugf(msg string, args ...interface{}) {
	l.logs.Debugf(msg, args...)
}

func (l logger) Error(msg string) {
	l.logs.Error(msg)
}

func (l logger) Infof(msg string, args ...interface{}) {
	l.logs.Infof(msg, args...)
}
