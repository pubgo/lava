package debug

import (
	"sort"
	"strings"

	"github.com/gofiber/fiber/v2"
	g "github.com/maragudk/gomponents"
	c "github.com/maragudk/gomponents/components"
	h "github.com/maragudk/gomponents/html"
	"github.com/pubgo/lava/core/debug"
)

func init() {
	initDebug()
}

func initDebug() {
	debug.Get("/", func(ctx *fiber.Ctx) error {
		pathMap := make(map[string]interface{})
		stack := debug.App().Stack()
		for m := range stack {
			for r := range stack[m] {
				route := stack[m][r]
				if strings.Contains(route.Path, "*") || strings.Contains(route.Path, ":") {
					continue
				}
				pathMap[route.Path] = nil
			}
		}

		var pathList []string
		for k := range pathMap {
			pathList = append(pathList, k)
		}
		sort.Strings(pathList)

		var nodes []g.Node
		nodes = append(nodes, h.H1(g.Text("routes")))
		for i := range pathList {
			path := "/debug" + pathList[i]
			nodes = append(nodes, h.A(g.Text(path), g.Attr("href", path)), h.Br())
		}
		ctx.Response().Header.SetContentType(fiber.MIMETextHTMLCharsetUTF8)
		return c.HTML5(c.HTML5Props{Title: "/app/routes", Body: nodes}).Render(ctx)
	})
}
