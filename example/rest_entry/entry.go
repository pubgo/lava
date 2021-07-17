package rest_entry

import (
	_ "expvar"
	"fmt"
	"net/http"

	"github.com/gofiber/fiber/v2"

	"github.com/pubgo/lug"
	"github.com/pubgo/lug/entry"
	"github.com/pubgo/lug/example/grpc_entry/handler"
)

var name = "test-http"

func GetEntry() entry.Entry {
	ent := lug.NewRest(name)
	ent.Description("entry http test")

	ent.Use(func(ctx *fiber.Ctx) error {
		fmt.Println("ok")
		return ctx.Next()
	})

	ent.BeforeStart(func() {
		go http.ListenAndServe(":8083", nil)
	})

	ent.Router(func(r fiber.Router) {
		r.Get("/", func(ctx *fiber.Ctx) error {
			_, err := ctx.WriteString("ok")
			return err
		})
	})

	ent.Register(handler.NewTestAPIHandler())

	return ent
}
