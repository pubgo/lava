package debug

import (
	"expvar"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/pubgo/lug/app"
	"github.com/pubgo/x/jsonx"
	"github.com/pubgo/xerror"
)

func init() {
	On(func(r *chi.Mux) {
		r.Handle("/debug/expvar", expvar.Handler())

		r.Get("/debug/vars", func(writer http.ResponseWriter, request *http.Request) {
			writer.Header().Set("Content-Type", "application/json")

			var keys []string
			expvar.Do(func(value expvar.KeyValue) {
				keys = append(keys, value.Key)
			})

			dt, err := jsonx.Marshal(keys)
			xerror.Panic(err)
			xerror.PanicErr(writer.Write(dt))
		})

		r.Get("/", func(writer http.ResponseWriter, request *http.Request) {
			writer.Header().Set("Content-Type", "application/json")

			var keys []string
			for _, r := range r.Routes() {
				keys = append(keys, fmt.Sprintf("http://localhost:%d%s", app.DebugPort, r.Pattern))
			}

			dt, err := jsonx.Marshal(keys)
			xerror.Panic(err)
			xerror.PanicErr(writer.Write(dt))
		})

		r.Get("/debug/vars/{vars}", func(writer http.ResponseWriter, request *http.Request) {
			vars := chi.URLParam(request, "vars")
			writer.Write([]byte(expvar.Get(vars).String()))
		})
	})
}
