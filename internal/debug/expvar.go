package debug

import (
	"expvar"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/pubgo/x/jsonx"
	"github.com/pubgo/x/strutil"
	"github.com/pubgo/xerror"
)

func init() {
	On(func(r *chi.Mux) {
		r.Handle("/debug/expvar", expvar.Handler())
		r.Get("/debug/vars", varsHandle)

		expvar.Do(func(kv expvar.KeyValue) {
			var val = kv.Value
			r.Get(fmt.Sprintf("/debug/vars/%s", kv.Key), func(writer http.ResponseWriter, request *http.Request) {
				writer.Header().Set("Content-Type", "application/json")
				xerror.PanicErr(writer.Write(strutil.ToBytes(val.String())))
			})
		})
	})
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
