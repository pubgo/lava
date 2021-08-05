package debug

import (
	"expvar"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/pubgo/x/jsonx"
	"github.com/pubgo/xerror"
)

func init() {
	On(func(r *chi.Mux) {
		r.Handle("/debug/expvar", expvar.Handler())
		r.Get("/debug/vars", varsHandle)
		r.Get("/debug/vars/{vars}", varsParams)
	})
}

func varsParams(writer http.ResponseWriter, request *http.Request) {
	vars := chi.URLParam(request, "vars")
	writer.Write([]byte(expvar.Get(vars).String()))
}

func varsHandle(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Content-Type", "application/json")

	var keys []string
	expvar.Do(func(value expvar.KeyValue) {
		keys = append(keys, value.Key)
	})

	dt, err := jsonx.Marshal(keys)
	xerror.Panic(err)
	xerror.PanicErr(writer.Write(dt))
}
