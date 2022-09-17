package varshandler

import (
	"expvar"
	"fmt"
	"net/http"

	"github.com/gofiber/adaptor/v2"
	"github.com/gofiber/fiber/v2"
	g "github.com/maragudk/gomponents"
	c "github.com/maragudk/gomponents/components"
	h "github.com/maragudk/gomponents/html"
	"github.com/pubgo/funk/recovery"
	"github.com/pubgo/lava/debug"
	"github.com/pubgo/xerror"
)

func init() {
	defer recovery.Exit()
	var index = func(keys []string) g.Node {
		var nodes []g.Node
		nodes = append(nodes, h.H1(g.Text("/expvar")))
		nodes = append(nodes, h.A(g.Text("/debug"), g.Attr("href", "/debug")), h.Br())
		for i := range keys {
			nodes = append(nodes, h.A(g.Text(keys[i]), g.Attr("href", keys[i])), h.Br())
		}

		return c.HTML5(c.HTML5Props{
			Title:    "/expvar",
			Language: "en",
			Body:     nodes,
		})
	}

	debug.Route("/expvar", func(r fiber.Router) {
		r.Get("/", adaptor.HTTPHandlerFunc(func(w http.ResponseWriter, request *http.Request) {
			var keys []string
			expvar.Do(func(kv expvar.KeyValue) {
				keys = append(keys, fmt.Sprintf("/debug/expvar/%s", kv.Key))
			})
			xerror.Panic(index(keys).Render(w))
		}))

		r.Get("/:name", func(ctx *fiber.Ctx) error {
			var name = ctx.Params("name")
			ctx.Response().Header.Set("Content-Type", "application/json; charset=utf-8")
			fmt.Fprintln(ctx, expvar.Get(name).String())
			return nil
		})
	})
}
