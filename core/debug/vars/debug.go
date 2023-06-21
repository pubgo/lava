package vars

import (
	"expvar"
	"fmt"

	"github.com/gofiber/fiber/v2"
	g "github.com/maragudk/gomponents"
	c "github.com/maragudk/gomponents/components"
	h "github.com/maragudk/gomponents/html"
	"github.com/pubgo/funk/recovery"

	"github.com/pubgo/lava/core/debug"
)

func init() {
	defer recovery.Exit()
	index := func(keys []string) g.Node {
		var nodes []g.Node
		nodes = append(nodes, h.H1(g.Text("/expvar")))
		nodes = append(nodes, h.A(g.Text("/debug"), g.Attr("href", "/debug")), h.Br())
		for i := range keys {
			nodes = append(nodes, h.A(g.Text(keys[i]), g.Attr("href", keys[i])), h.Br())
		}
		return c.HTML5(c.HTML5Props{Title: "/expvar", Body: nodes})
	}

	debug.Route("/vars", func(r fiber.Router) {
		r.Get("/", func(ctx *fiber.Ctx) error {
			ctx.Response().Header.SetContentType(fiber.MIMETextHTMLCharsetUTF8)
			var keys []string
			expvar.Do(func(kv expvar.KeyValue) {
				keys = append(keys, fmt.Sprintf("/debug/vars/%s", kv.Key))
			})

			return index(keys).Render(ctx)
		})

		r.Get("/:name", func(ctx *fiber.Ctx) error {
			name := ctx.Params("name")
			ctx.Response().Header.Set("Content-Type", "application/json; charset=utf-8")
			return ctx.SendString(expvar.Get(name).String())
		})
	})
}
