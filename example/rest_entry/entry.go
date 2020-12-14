package rest_entry

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/pubgo/golug"
	"github.com/pubgo/golug/golug_entry"
)

func GetEntry() golug_entry.Entry {
	ent := golug.NewRestEntry("http")
	ent.Version("v0.0.1")
	ent.Description("entry http test")

	ent.Use(func(ctx *fiber.Ctx) error {
		fmt.Println("ok")

		return ctx.Next()
	})

	ent.Router("/", func(r fiber.Router) {
		r.Get("/", func(ctx *fiber.Ctx) error {
			_, err := ctx.WriteString("ok")
			return err
		})
	})

	return ent
}
