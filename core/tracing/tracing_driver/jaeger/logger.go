package jaeger

import (
	"github.com/pubgo/lava/logging"
	jLog "github.com/uber/jaeger-client-go/log"
	"go.uber.org/zap"
)

var _ jLog.Logger = (*traceLog)(nil)

func newLog(name string) *traceLog {
	return &traceLog{logs: logging.Component(name).Depth(2)}
}

type traceLog struct {
	logs *zap.Logger
}

func (l traceLog) Write(p []byte) (n int, err error) {
	l.logs.Info(string(p))
	return 0, err
}

func (l traceLog) Debugf(msg string, args ...interface{}) {
	l.logs.Sugar().Debugf(msg, args...)
}

func (l traceLog) Error(msg string) {
	l.logs.Error(msg)
}

func (l traceLog) Infof(msg string, args ...interface{}) {
	l.logs.Sugar().Infof(msg, args...)
}
