package swagger

import (
	"encoding/json"
	"fmt"
	"github.com/pubgo/lava/core/debug"
	"github.com/pubgo/lava/pkg/syncx"
	"html/template"
	"io/fs"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/go-openapi/loads"
	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/spec"
	"github.com/pkg/browser"
	"github.com/pubgo/xerror"
	httpSwagger "github.com/swaggo/http-swagger"
	"github.com/urfave/cli/v2"

	"github.com/pubgo/lava/pkg/clix"
)

var Cmd = &cli.Command{
	Name:        "swagger",
	Usage:       "start swagger web",
	Description: clix.ExampleFmt(`lava rest.http`),
	Action: func(ctx *cli.Context) error {
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

		debug.Get("/", debug.WrapFunc(func(writer http.ResponseWriter, request *http.Request) {
			var names []string
			xerror.Panic(filepath.Walk("./docs", func(path string, info fs.FileInfo, err error) error {
				if strings.HasSuffix(path, ".swagger.json") {
					names = append(names, strings.TrimPrefix(path, "docs/"))
				}
				return nil
			}))
			xerror.Panic(homeTmpl.Execute(writer, names))
		}))

		debug.Get("/docs/*", debug.WrapFunc(func(writer http.ResponseWriter, request *http.Request) {
			writer.Header().Set("Content-Type", "application/json")
			var bytes, err = ioutil.ReadFile(strings.Trim(request.RequestURI, "/"))
			xerror.Panic(err)

			specDoc, err := loads.Analyzed(bytes, "")
			xerror.Panic(err)

			specDoc, err = specDoc.Expanded(&spec.ExpandOptions{
				SkipSchemas:         false,
				ContinueOnError:     true,
				AbsoluteCircularRef: true,
			})
			xerror.Panic(err)

			b, err := json.MarshalIndent(specDoc.Spec(), "", "  ")
			xerror.Panic(err)
			writer.Write(b)
		}))

		debug.Get("/swagger/*", debug.WrapFunc(func(writer http.ResponseWriter, request *http.Request) {
			var s ServeCmd
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
					SpecURL:  fmt.Sprintf("/docs/%s", request.URL.Path),
					Path:     request.URL.Path,
				}, handler)
			} else if flavor == "swagger" {
				handler = middleware.SwaggerUI(middleware.SwaggerUIOpts{
					Title:    request.URL.Path,
					BasePath: basePath,
					SpecURL:  fmt.Sprintf("/docs/%s", request.URL.Path),
					Path:     request.URL.Path,
				}, handler)
			} else {
				handler = httpSwagger.Handler(
					httpSwagger.URL(fmt.Sprintf("%s.swagger.json", request.URL.Path)),
				)
			}
			handler.ServeHTTP(writer, request)
		}))

		syncx.GoDelay(func() {
			xerror.Panic(browser.OpenURL("http://localhost:8082"))
		})
		return nil
	},
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
