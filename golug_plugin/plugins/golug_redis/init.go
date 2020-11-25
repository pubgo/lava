package golug_redis

import (
	"github.com/go-redis/redis/v7"
	"github.com/pubgo/golug/golug_log"
	"github.com/pubgo/xerror"
	"github.com/pubgo/xlog"
)

var options *redis.Options
var name = "redis"
var log xlog.XLog

func init() {
	xerror.Exit(golug_log.Watch(func(logs xlog.XLog) { log = logs.Named(name) }))
}
