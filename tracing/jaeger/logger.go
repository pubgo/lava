package jaeger

import (
	"github.com/pubgo/xlog"
	jaegerLog "github.com/uber/jaeger-client-go/log"
)

var _ jaegerLog.Logger = (*logger)(nil)

type logger struct {
	logs xlog.Xlog
}

func (l logger) Error(msg string) {
	l.logs.Error(msg)
}

func (l logger) Infof(msg string, args ...interface{}) {
	l.logs.Infof(msg, args...)
}
