package golug_consul

import (
	"sync"
	"time"

	"github.com/pubgo/golug/golug_log"
	"github.com/pubgo/xlog"
)

var name = "consul"
var clientM sync.Map
var log xlog.XLog

const Timeout = time.Second * 2

func init() {
	golug_log.Watch(func(logs xlog.XLog) { log = logs.Named(name) })
}
