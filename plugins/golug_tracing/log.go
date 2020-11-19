package golug_tracing

import (
	"github.com/pubgo/golug/golug_log"
	"github.com/pubgo/xerror"
	"github.com/pubgo/xlog"
)

var log xlog.XLog

func init() {
	xerror.Panic(golug_log.Watch(func(logs xlog.XLog) {
		log = logs.Named(name)
	}))
}

type tracingLogger struct{}

func (l *tracingLogger) Error(msg string) {
	log.Error(msg)
}

func (l *tracingLogger) Infof(msg string, args ...interface{}) {
	log.Infof(msg, args...)
}
