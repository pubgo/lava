package entry

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/pubgo/golug"
	"github.com/pubgo/golug/golug_entry"
	"github.com/pubgo/xerror"
)

func GetEntry() golug_entry.Entry {
	ent := golug.NewEntry("entry")
	xerror.Panic(ent.Version("v0.0.1"))
	xerror.Panic(ent.Description("entry http test"))

	ent.Use(func(ctx *fiber.Ctx) error {
		fmt.Println("ok")

		return ctx.Next()
	})

	ent.Group("/api", func(r fiber.Router) {
		r.Get("/", func(ctx *fiber.Ctx) error {
			_, err := ctx.WriteString("ok")
			return err
		})
	})

	return ent
}
