package debug

import (
	"fmt"
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

		dt, err := jsonx.Marshal(keys)
		xerror.Panic(err)
		xerror.PanicErr(writer.Write(dt))
	}
}
