package statsviz

import (
	"strings"

	"github.com/arl/statsviz"
	"github.com/gofiber/fiber/v2"
	"github.com/pubgo/funk/assert"
	"github.com/pubgo/lava/core/debug"
	"github.com/pubgo/lava/pkg/httputil"
)

func init() {
	srv := assert.Exit1(statsviz.NewServer(statsviz.Root("/statsviz")))
	debug.Route("/statsviz", func(router fiber.Router) {
		router.Use(func(ctx *fiber.Ctx) error {
			path := string(ctx.Request().URI().Path())
			pathList := strings.Split(path, "/")
			if strings.Trim(pathList[len(pathList)-1], "/") == "ws" {
				return httputil.HTTPHandler(srv.Ws())(ctx)
			}

			return ctx.Next()
		})

		router.Get("/", func(ctx *fiber.Ctx) error {
			return httputil.HTTPHandler(srv.Index())(ctx)
		})
		router.Get("/*", func(ctx *fiber.Ctx) error {
			return httputil.HTTPHandler(srv.Index())(ctx)
		})
	})
}
