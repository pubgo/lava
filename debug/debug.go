package debug

import (
	"fmt"
	"sort"
	"strings"

	"github.com/gofiber/fiber/v2"
	g "github.com/maragudk/gomponents"
	c "github.com/maragudk/gomponents/components"
	h "github.com/maragudk/gomponents/html"
	"github.com/pubgo/funk/recovery"
)

func init() {
	initDebug()
}

func initDebug() {
	defer recovery.Exit()
	Get("/", func(ctx *fiber.Ctx) error {
		var pathMap = make(map[string]interface{})
		stack := App().Stack()
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
			k = strings.TrimRight(k, "/") + "/"

			if strings.Contains(strings.Trim(k, "/"), "/") {
				continue
			}

			pathList = append(pathList, fmt.Sprintf("/debug%s", k))
		}
		sort.Strings(pathList)

		var nodes []g.Node
		nodes = append(nodes, h.H1(g.Text("routes")))
		for i := range pathList {
			nodes = append(nodes, h.A(g.Text(pathList[i]), g.Attr("href", pathList[i])), h.Br())
		}
		ctx.Response().Header.SetContentType(fiber.MIMETextHTMLCharsetUTF8)
		return c.HTML5(c.HTML5Props{Title: "/app/routes", Body: nodes}).Render(ctx)
	})
}
