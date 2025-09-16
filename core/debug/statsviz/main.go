package statsviz

import (
	"strings"

	"github.com/arl/statsviz"
	"github.com/gofiber/fiber/v2"
	"github.com/pubgo/funk/v2"
	"github.com/pubgo/funk/v2/assert"
	"github.com/pubgo/lava/v2/core/debug"
	"github.com/pubgo/lava/v2/pkg/httputil"
)

// github.com/go-echarts/statsview

func init() {
	srv := assert.Exit1(statsviz.NewServer(statsviz.Root("/debug/statsviz")))
	debug.Route("/statsviz", func(router fiber.Router) {
		router.Use(func(ctx *fiber.Ctx) error {
			path := string(ctx.Request().URI().Path())
			lastPath := strings.TrimSpace(funk.Last(strings.Split(strings.Trim(path, "/"), "/")))

			if lastPath == "ws" {
				return httputil.HTTPHandler(srv.Ws())(ctx)
			}

			if lastPath == "statsviz" {
				return httputil.HTTPHandler(srv.Index())(ctx)
			}

			return ctx.Next()
		})

		router.Get("", func(ctx *fiber.Ctx) error {
			return httputil.HTTPHandler(srv.Index())(ctx)
		})
		router.Get("/", func(ctx *fiber.Ctx) error {
			return httputil.HTTPHandler(srv.Index())(ctx)
		})
		router.Get("/*", func(ctx *fiber.Ctx) error {
			return httputil.HTTPHandler(srv.Index())(ctx)
		})
	})
}
