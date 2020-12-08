package grpc_entry

import (
	"github.com/pubgo/golug/golug_log"
	"github.com/pubgo/xlog"
)

const Name = "grpcEntry"

var log xlog.XLog

func init() {
	golug_log.Watch(func(logs xlog.XLog) { log = logs.Named(Name) })
}
