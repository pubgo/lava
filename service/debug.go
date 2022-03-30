package service

import (
	"html/template"
	"strings"

	"github.com/gofiber/fiber/v2"

	"github.com/pubgo/lava/pkg/htmlx"
)

func (t *serviceImpl) initDebug() {
	var homeTmpl = template.Must(template.New("index").Parse(`
		<html>
		<head>
		<title>/app/routes</title>
		</head>
		<body>
		{{range .}}
			<a href={{.}}>{{.}}</a><br/>
		{{end}}
		</body>
		</html>
		`))

	var handler = func(ctx *fiber.Ctx) error {
		var keys []string
		stack := t.gw.Get().Stack()
		for m := range stack {
			for r := range stack[m] {
				route := stack[m][r]
				keys = append(keys, strings.TrimSuffix(route.Path, "*"))
			}
		}

		return htmlx.Html(ctx, homeTmpl, keys)
	}

	t.app.Get("/debug", handler)
}
