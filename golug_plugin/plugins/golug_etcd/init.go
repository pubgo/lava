package golug_etcd

import (
	"sync"
	"time"

	"github.com/pubgo/golug/golug_log"
	"github.com/pubgo/xerror"
	"github.com/pubgo/xlog"
)

var name = "etcd"
var log xlog.XLog
var cfg Cfg
var clientM sync.Map

const Timeout = time.Second * 2

func init() {
	xerror.Panic(golug_log.Watch(func(logs xlog.XLog) { log = logs.Named(name) }))
}
