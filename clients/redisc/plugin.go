package redisc

import (
	"github.com/go-redis/redis/v8"
	"github.com/pubgo/xerror"

	"github.com/pubgo/lava/config"
	"github.com/pubgo/lava/pkg/ctxutil"
	"github.com/pubgo/lava/pkg/merge"
	"github.com/pubgo/lava/plugin"
	"github.com/pubgo/lava/resource"
)

func init() {
	plugin.Register(&plugin.Base{
		Name: Name,
		OnInit: func(p plugin.Process) {
			xerror.Panic(config.Decode(Name, &cfgMap))
			for k, v := range cfgMap {
				client := redis.NewClient(merge.Struct(DefaultCfg(), v).(*redis.Options))
				xerror.PanicF(client.Ping(ctxutil.Timeout()).Err(), "redis(%s)连接失败", k)
				resource.Update(k, &Client{cli: client})
			}
		},
	})
}
