package task_entry

import (
	"github.com/pubgo/golug/golug_log"
	"github.com/pubgo/xlog"
)

const Name = "task_entry"

var log xlog.XLog

func init() {
	golug_log.Watch(func(logs xlog.XLog) {
		log = logs.Named(Name)
	})
}
