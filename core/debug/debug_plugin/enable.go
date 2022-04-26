package debug_plugin

import (
	"fmt"
	"github.com/pubgo/lava/logging/logutil"
	"sort"
	"strings"

	pongo "github.com/flosch/pongo2/v5"
	"github.com/gofiber/fiber/v2"
	"github.com/pkg/browser"
	"github.com/pubgo/xerror"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"

	"github.com/pubgo/lava/core/debug"
	"github.com/pubgo/lava/pkg/htmlx"
	"github.com/pubgo/lava/pkg/netutil"
	"github.com/pubgo/lava/pkg/syncx"
	"github.com/pubgo/lava/service"
)

func Enable(srv service.Service) {
	srv.RegisterApp("/debug", debug.App())
	initDebug()

	var openWeb bool
	srv.Flags(&cli.BoolFlag{
		Name:        "debug.web",
		Value:       openWeb,
		Destination: &openWeb,
		Usage:       "open web browser",
	})

	srv.AfterStarts(func() {
		if !openWeb {
			return
		}

		syncx.GoSafe(func() {
			logutil.ErrRecord(zap.L(),
				browser.OpenURL(fmt.Sprintf("http://%s:%d/debug", netutil.GetLocalIP(), srv.Options().Port)))
		})
	})
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

	debug.App().Get("/", func(ctx *fiber.Ctx) error {
		var pathMap = make(map[string]interface{})
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
