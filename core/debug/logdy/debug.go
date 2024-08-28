package logdy

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/logdyhq/logdy-core/logdy"
	_ "github.com/logdyhq/logdy-core/logdy"
	"github.com/pubgo/lava/core/debug"
	"github.com/pubgo/lava/pkg/httputil"
)

func init() {
	mux := http.NewServeMux()
	logdy.InitializeLogdy(logdy.Config{HttpPathPrefix: "/logs"}, mux)

	debug.Route("/logs", func(r fiber.Router) {
		//r.Use(func(ctx *fiber.Ctx) error {
		//	path := string(ctx.Request().URI().Path())
		//	fmt.Println(path)
		//
		//	return ctx.Next()
		//})

		r.Get("/", func(ctx *fiber.Ctx) error {
			return httputil.HTTPHandler(mux)(ctx)
		})
		r.Get("/*", func(ctx *fiber.Ctx) error {
			return httputil.HTTPHandler(mux)(ctx)
		})
	})
}
