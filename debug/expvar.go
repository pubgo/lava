package debug

import (
	"expvar"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/pubgo/lug/app"
)

func init() {
	On(func(r *chi.Mux) {
		r.Handle("/debug/expvar", expvar.Handler())

		r.Get("/debug/vars", func(writer http.ResponseWriter, request *http.Request) {
			var dt = ""
			expvar.Do(func(value expvar.KeyValue) {
				dt += value.Key + "\n"
			})
			writer.Write([]byte(dt))
		})
		r.Get("/", func(writer http.ResponseWriter, request *http.Request) {
			var dt = ""
			for _, r := range r.Routes() {
				dt += fmt.Sprintf("http://localhost:%d%s\n", app.DebugPort, r.Pattern)
			}

			writer.Write([]byte(dt))
		})

		r.Get("/debug/vars/{vars}", func(writer http.ResponseWriter, request *http.Request) {
			vars := chi.URLParam(request, "vars")
			writer.Write([]byte(expvar.Get(vars).String()))
		})
	})
}
