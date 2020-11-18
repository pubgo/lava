package golug_redis

import (
	"github.com/go-redis/redis"
	"github.com/gofiber/fiber/v2"
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
			ent.Use(func(ctx *fiber.Ctx) error {

			})

		},
	}))
}
