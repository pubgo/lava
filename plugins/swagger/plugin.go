package swagger

import (
	"encoding/json"
	"net/http"
	"path"

	"github.com/go-openapi/loads"
	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/spec"
	"github.com/pubgo/xerror"

	"github.com/pubgo/lug/mux"
)

func init() {
	mux.Get("/swagger/*", func(writer http.ResponseWriter, request *http.Request) {
		var s ServeCmd
		specDoc, err := loads.Analyzed(json.RawMessage(nil), "")
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

		basePath := s.BasePath
		if basePath == "" {
			basePath = "/"
		}

		visit := s.DocURL
		handler := http.NotFoundHandler()
		if !s.NoUI {
			if s.Flavor == "redoc" {
				handler = middleware.Redoc(middleware.RedocOpts{
					BasePath: basePath,
					SpecURL:  path.Join(basePath, "swagger.json"),
					Path:     s.Path,
				}, handler)
			} else if visit != "" || s.Flavor == "swagger" {
				handler = middleware.SwaggerUI(middleware.SwaggerUIOpts{
					BasePath: basePath,
					SpecURL:  path.Join(basePath, "swagger.json"),
					Path:     s.Path,
				}, handler)
			}
		}
		middleware.Spec(basePath, b, handler).ServeHTTP(writer, request)
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
