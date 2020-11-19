package golug_redis

import (
	"github.com/go-redis/redis/v7"
	"github.com/pubgo/golug/golug_entry"
	"github.com/pubgo/golug/golug_plugin"
	"github.com/pubgo/xerror"
)

var options *redis.Options
var name = "redis"

func init() {
	xerror.Exit(golug_plugin.Register(&golug_plugin.Base{
		Name: name,
		OnInit: func(ent golug_entry.Entry) {
		},
	}))
}
