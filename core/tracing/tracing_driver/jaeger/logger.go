package jaeger

import (
	"github.com/pubgo/funk/log"
	jLog "github.com/uber/jaeger-client-go/log"
)

var _ jLog.Logger = (*traceLog)(nil)

func newLog(name string) *traceLog {
	return &traceLog{logs: log.GetLogger(name).WithCallerSkip(2)}
}

type traceLog struct {
	logs log.Logger
}

func (l traceLog) Write(p []byte) (n int, err error) {
	l.logs.Info().Msg(string(p))
	return 0, err
}

func (l traceLog) Debugf(msg string, args ...interface{}) {
	l.logs.Debug().Msgf(msg, args...)
}

func (l traceLog) Error(msg string) {
	l.logs.Error().Msg(msg)
}

func (l traceLog) Infof(msg string, args ...interface{}) {
	l.logs.Info().Msgf(msg, args...)
}
