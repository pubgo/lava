package rest_entry

import (
	"fmt"
	"net"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/pubgo/lug"
	"github.com/pubgo/lug/entry"
	"github.com/pubgo/xerror"
)

var name = "test-http"

func GetEntry() entry.Abc {
	ent := lug.NewRest(name)
	ent.Version("v0.0.1")
	ent.Description("entry http test")

	ent.Use(func(ctx *fiber.Ctx) error {
		fmt.Println("ok")

		return ctx.Next()
	})

	ent.BeforeStart(func() {
		l, err := net.Listen("tcp", ":8083")
		xerror.Panic(err)
		go http.Serve(l, nil)
	})

	ent.Router(func(r fiber.Router) {
		r.Get("/", func(ctx *fiber.Ctx) error {
			_, err := ctx.WriteString("ok")
			return err
		})
	})

	return ent
}
