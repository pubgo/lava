package debug

import (
	"github.com/pubgo/lava/mux"
	"html/template"
	"net/http"
	"strings"

	"github.com/pubgo/xerror"
)

func init() {
	http.HandleFunc("/", home())
	http.Handle("/debug", http.RedirectHandler("/", http.StatusTemporaryRedirect))
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
			keys = append(keys, strings.TrimSuffix(r.Pattern, "/*"))
		}
		xerror.Panic(homeTmpl.Execute(writer, keys))
	}
}
