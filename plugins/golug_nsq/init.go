package golug_nsq

import (
	"github.com/pubgo/golug/golug_log"
	"github.com/pubgo/xerror"
	"github.com/pubgo/xlog"
)

var name = "nsq"
var log xlog.XLog

func init() {
	xerror.Panic(golug_log.Watch(func(logs xlog.XLog) { log = logs.Named(name) }))
}
