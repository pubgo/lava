package swagger

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strings"

	"github.com/go-openapi/loads"
	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/spec"
	"github.com/pubgo/xerror"
	httpSwagger "github.com/swaggo/http-swagger"

	"github.com/pubgo/lug/mux"
)

func Init(names func() []string, asset func(name string) []byte) {
	var homeTmpl = template.Must(template.New("index").Parse(`
		<html>
		<head>
		<title>/swagger</title>
		</head>
		<body>
		{{range .}}
			<a href={{.}}>{{.}}</a><br/>
		{{end}}
		</body>
		</html>
		`))
	mux.Get("/swagger", func(writer http.ResponseWriter, request *http.Request) {
		var keys []string
		for _, r := range names() {
			keys = append(keys, strings.TrimSuffix(r, ".swagger.json"))
		}
		xerror.Panic(homeTmpl.Execute(writer, keys))
	})

	mux.Get("/swagger/*", func(writer http.ResponseWriter, request *http.Request) {
		var s ServeCmd

		if strings.HasSuffix(request.RequestURI, "swagger.json") {
			writer.Header().Set("Content-Type", "application/json")

			specDoc, err := loads.Analyzed(asset(strings.Trim(request.RequestURI, "/")), "")
			xerror.Panic(err)

			if s.Flatten {
				specDoc, err = specDoc.Expanded(&spec.ExpandOptions{
					SkipSchemas:         false,
					ContinueOnError:     true,
					AbsoluteCircularRef: true,
				})
				xerror.Panic(err)
			}

			b, err := json.MarshalIndent(specDoc.Spec(), "", "  ")
			xerror.Panic(err)
			writer.Write(b)
			return
		}

		var flavor = "swagger"
		if f := request.URL.Query().Get("flavor"); f != "" {
			flavor = f
		}

		basePath := s.BasePath
		if basePath == "" {
			basePath = "/"
		}

		handler := http.NotFoundHandler()
		if flavor == "redoc" {
			handler = middleware.Redoc(middleware.RedocOpts{
				Title:    request.URL.Path,
				BasePath: basePath,
				SpecURL:  fmt.Sprintf("%s.swagger.json", request.URL.Path),
				Path:     request.URL.Path,
			}, handler)
		} else if flavor == "swagger" {
			handler = middleware.SwaggerUI(middleware.SwaggerUIOpts{
				Title:    request.URL.Path,
				BasePath: basePath,
				SpecURL:  fmt.Sprintf("%s.swagger.json", request.URL.Path),
				Path:     request.URL.Path,
			}, handler)
		} else {
			handler = httpSwagger.Handler(
				httpSwagger.URL(fmt.Sprintf("%s.swagger.json", request.URL.Path)),
			)
		}
		handler.ServeHTTP(writer, request)
	})
}

// ServeCmd to serve a swagger spec with docs ui
type ServeCmd struct {
	BasePath string `long:"base-path" description:"the base path to serve the spec and UI at"`
	Flavor   string `short:"F" long:"flavor" description:"the flavor of docs, can be swagger or redoc" default:"redoc" choice:"redoc" choice:"swagger"`
	DocURL   string `long:"doc-url" description:"override the url which takes a url query param to render the doc ui"`
	NoOpen   bool   `long:"no-open" description:"when present won't open the the browser to show the url"`
	NoUI     bool   `long:"no-ui" description:"when present, only the swagger spec will be served"`
	Flatten  bool   `long:"flatten" description:"when present, flatten the swagger spec before serving it"`
	Port     int    `long:"port" short:"p" description:"the port to serve this site" env:"PORT"`
	Host     string `long:"host" description:"the interface to serve this site, defaults to 0.0.0.0" default:"0.0.0.0" env:"HOST"`
	Path     string `long:"path" description:"the uri path at which the docs will be served" default:"docs"`
}
