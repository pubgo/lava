package debug

import (
	"fmt"
	"html/template"
	"net/http"
	"sort"

	"github.com/pubgo/xerror"
)

func init() {
	http.HandleFunc("/", home())
}

func home() func(writer http.ResponseWriter, r *http.Request) {
	type RouteInfo struct {
		Method  string
		Path    string
		Handler string
	}

	return func(writer http.ResponseWriter, req *http.Request) {
		serveMux.mu.RLock()
		defer serveMux.mu.RUnlock()

		var paths []RouteInfo
		for k, v := range serveMux.m {
			paths = append(paths, RouteInfo{
				Method:  "any",
				Path:    k,
				Handler: fmt.Sprintf("%#v", v),
			})
		}

		sort.Slice(paths, func(i, j int) bool { return paths[i].Path < paths[j].Path })
		xerror.Panic(indexTmpl.Execute(writer, paths))
	}
}

var indexTmpl = template.Must(template.New("index").Parse(`
<html>
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
