package golug_nsq

import (
	"sync"

	"github.com/pubgo/golug/golug_log"
	"github.com/pubgo/xerror"
	"github.com/pubgo/xlog"
)

var name = "nsq"
var log xlog.XLog
var cfg Cfg
var nsqM sync.Map

func init() {
	xerror.Panic(golug_log.Watch(func(logs xlog.XLog) { log = logs.Named(name) }))
}
