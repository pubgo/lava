package golug_request_id

import (
	"github.com/gofiber/fiber/v2"
	"github.com/hashicorp/go-uuid"
	"github.com/pubgo/golug/golug_entry"
	"github.com/pubgo/golug/golug_plugin"
	"github.com/pubgo/xerror"
)

const name = "request_id"
const RequestId = "X-Request-Id"

func init() {
	xerror.Exit(golug_plugin.Register(&golug_plugin.Base{
		Enabled: true,
		Name:    name,
		OnInit: func(ent golug_entry.Entry) {
			xerror.Panic(ent.UnWrap(func(entry golug_entry.HttpEntry) {
				entry.Use(func(ctx *fiber.Ctx) error {
					rid := ctx.Get(RequestId, xerror.PanicStr(uuid.GenerateUUID()))
					ctx.Set(RequestId, rid)
					return xerror.Wrap(ctx.Next())
				})
			}))
		},
	}))
}
