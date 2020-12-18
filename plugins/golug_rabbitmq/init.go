package golug_rabbitmq

import (
	"github.com/pubgo/golug/golug_log"
	"github.com/pubgo/xlog"
)

var name = "rabbitmq"
var log xlog.XLog

func init() {
	golug_log.Watch(func(logs xlog.XLog) { log = logs.Named(name) })
}
