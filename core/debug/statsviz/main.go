package statsviz

import (
	"github.com/arl/statsviz"
	"github.com/gofiber/fiber/v3"
	"github.com/pubgo/funk/assert"
	"github.com/pubgo/lava/core/debug"
	"github.com/pubgo/lava/pkg/httputil"
	"strings"
)

func init() {
	srv := assert.Exit1(statsviz.NewServer())
	debug.Group("/statsviz", func(ctx fiber.Ctx) error {
		path := string(ctx.Request().URI().Path())
		pathList := strings.Split(path, "/")
		if strings.Trim(pathList[len(pathList)-1], "/") == "ws" {
			return httputil.HTTPHandler(srv.Ws())(ctx)
		}

		return httputil.HTTPHandler(srv.Index())(ctx)
	})
}
