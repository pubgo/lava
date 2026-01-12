package debug

import (
	"path/filepath"
	"sort"

	"github.com/gofiber/fiber/v2"
	g "github.com/maragudk/gomponents"
	c "github.com/maragudk/gomponents/components"
	h "github.com/maragudk/gomponents/html"

	"github.com/pubgo/lava/v2/core/debug"
)

func init() {
	initDebug()
}

func initDebug() {
	debug.Get("/", func(ctx *fiber.Ctx) error {
		pathMap := make(map[string]any)

		stack := ctx.App().Stack()
		for m := range stack {
			for r := range stack[m] {
				route := stack[m][r]
				//if strings.Contains(route.Path, "*") || strings.Contains(route.Path, ":") {
				//	continue
				//}
				pathMap[route.Path] = nil
			}
		}

		pathList := make([]string, 0, len(pathMap))
		for k := range pathMap {
			pathList = append(pathList, k)
		}
		sort.Strings(pathList)

		nodes := make([]g.Node, 0, len(pathList))
		nodes = append(nodes, h.H1(g.Text("routes")))
		for i := range pathList {
			path := pathList[i]
			nodes = append(nodes, h.A(g.Text(path), g.Attr("href", filepath.Join(string(ctx.Request().Header.Peek("Path-Prefix")), path))), h.Br())
		}
		ctx.Response().Header.SetContentType(fiber.MIMETextHTMLCharsetUTF8)
		return c.HTML5(c.HTML5Props{Title: "/app/routes", Body: nodes}).Render(ctx)
	})
}
