package golug_timeout

import (
	"github.com/pubgo/golug/golug_log"
	"github.com/pubgo/xlog"
)

var name = "timeout"
var log xlog.XLog

func init() {
	golug_log.Watch(func(logs xlog.XLog) { log = logs.Named(name) })
}
