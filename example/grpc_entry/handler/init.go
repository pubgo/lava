package handler

import (
	"github.com/pubgo/golug/golug_log"
	"github.com/pubgo/xerror"
	"github.com/pubgo/xlog"
)

var log xlog.XLog

func init() {
	xerror.Exit(golug_log.Watch(func(logs xlog.XLog) {
		log = logs.Named("hello.handler")
	}))
}
