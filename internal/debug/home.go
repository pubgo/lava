package debug

import (
	"fmt"
	"html/template"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/pubgo/x/jsonx"
	"github.com/pubgo/xerror"
)

func init() {
	On(func(r *chi.Mux) {
		r.Get("/", home(r))
	})
}

func home(r *chi.Mux) func(writer http.ResponseWriter, r *http.Request) {
	return func(writer http.ResponseWriter, req *http.Request) {
		writer.Header().Set("Content-Type", "application/json")

		var keys []string
		for _, r := range r.Routes() {
			keys = append(keys, fmt.Sprintf("http://localhost%s%s", Addr, r.Pattern))
		}

		//xerror.Panic(indexTmpl.Execute(writer, paths))
		dt, err := jsonx.Marshal(keys)
		xerror.Panic(err)
		xerror.PanicErr(writer.Write(dt))
	}
}

var indexTmpl = template.Must(template.New("index").Parse(`<html>
<head>
<title>/debug/routes</title>
</head>
<body>
<table>
<thead><td>Method</td><td>Path</td><td>Handler</td></thead>
{{range .}}
	<tr>
	<td>{{.Method}}</td><td><a href={{.Path}}>{{.Path}}</a></td><td>{{.Handler}}</td>
	</tr>
{{end}}
</table>
</body>
</html>
`))