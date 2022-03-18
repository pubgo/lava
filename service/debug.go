package service

import (
	"github.com/gofiber/fiber/v2"
	"github.com/pubgo/xerror"
	"html/template"
	"strings"
)

func init() {
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
	Get("/", func(ctx *fiber.Ctx) error {
		var keys []string
		for _, r := range Mux().Routes() {
			keys = append(keys, strings.TrimSuffix(r.Pattern, "*"))
		}
		xerror.Panic(homeTmpl.Execute(writer, keys))
	})
}
