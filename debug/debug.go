package debug

import (
	"fmt"
	"sort"
	"strings"

	pongo "github.com/flosch/pongo2/v5"
	"github.com/gofiber/fiber/v2"
	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/recovery"

	"github.com/pubgo/lava/pkg/htmlx"
)

func init() {
	initDebug()
}

func initDebug() {
	defer recovery.Exit()
	temp := pongo.Must(pongo.FromString(strings.TrimSpace(`
	<html>
		<head>
		<title>/app/routes</title>
		</head>
		<body>
 		{% for path in data %}
			<a href={{path}}>{{path}}</a><br/>
		{% endfor %}
		</body>
	</html>	
`)))

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

		var data = assert.Must1(temp.ExecuteBytes(htmlx.Context{"data": pathList}))
		return htmlx.Html(ctx, data)
	})
}
