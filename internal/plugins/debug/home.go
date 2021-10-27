package debug

import (
	"html/template"
	"net/http"
	"strings"

	"github.com/pubgo/xerror"

	"github.com/pubgo/lava/mux"
)

func init() {
	mux.Get("/", home())
	mux.Get("/debug", home())
}

func home() func(writer http.ResponseWriter, r *http.Request) {
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

	return func(writer http.ResponseWriter, req *http.Request) {
		var keys []string
		for _, r := range mux.Mux().Routes() {
			keys = append(keys, strings.TrimSuffix(r.Pattern, "*"))
		}
		xerror.Panic(homeTmpl.Execute(writer, keys))
	}
}
