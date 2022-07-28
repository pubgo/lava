package debug

import (
	"fmt"
	"github.com/pubgo/funk/recovery"
	"sort"
	"strings"

	pongo "github.com/flosch/pongo2/v5"
	"github.com/gofiber/fiber/v2"
	"github.com/pubgo/xerror"

	"github.com/pubgo/lava/internal/pkg/htmlx"
)

func init() {
	defer recovery.Exit()

	initDebug()
}

func initDebug() {
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

		var data, err = temp.ExecuteBytes(htmlx.Context{"data": pathList})
		xerror.Panic(err)
		return htmlx.Html(ctx, data)
	})
}
