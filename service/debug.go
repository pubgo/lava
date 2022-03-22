package service

import (
	"html/template"
	"strings"

	"github.com/gofiber/fiber/v2"
)

func (t *implService) initDebug() {
	var homeTmpl = template.Must(template.New("index").Parse(`
		<html>
		<head>
		<title>/debug/routes</title>
		</head>
		<body>
		{{range .}}
			<a href={{.}}>{{.}}</a><br/>
		{{end}}
		</body>
		</html>
		`))

	t.Debug().Get("/", func(ctx *fiber.Ctx) error {
		var keys []string
		stack := t.gw.Get().Stack()
		for m := range stack {
			for r := range stack[m] {
				route := stack[m][r]
				keys = append(keys, strings.TrimSuffix(route.Path, "*"))
			}
		}

		return homeTmpl.Execute(ctx, keys)
	})
}
