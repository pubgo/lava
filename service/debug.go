package service

import (
	"bytes"
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

		return html(ctx, homeTmpl, keys)
	})
}

func html(ctx *fiber.Ctx, temp *template.Template, data any) error {
	if data == nil {
		data = map[string]interface{}{}
	}
	var buf = bytes.NewBuffer(nil)
	if err := temp.Execute(buf, data); err != nil {
		return err
	}
	ctx.Response().Header.SetContentType(fiber.MIMETextHTMLCharsetUTF8)
	ctx.Response().SetBody(buf.Bytes())
	return nil
}
