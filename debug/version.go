package debug

import (
	"github.com/go-chi/chi/v5"
	"github.com/pubgo/lug/version"
	"github.com/pubgo/x/jsonx"
	"github.com/pubgo/xerror"
	"os"
	"runtime/debug"

	"net/http"
)

func init() {
	On(func(mux *chi.Mux) {
		mux.Get("/env", func(writer http.ResponseWriter, request *http.Request) {
			writer.Header().Set("Content-Type", "application/json")

			dt, err := jsonx.Marshal(os.Environ())
			xerror.Panic(err)
			xerror.PanicErr(writer.Write(dt))
		})

		mux.Get("/version", func(writer http.ResponseWriter, request *http.Request) {
			writer.Header().Set("Content-Type", "application/json")

			dt, err := jsonx.Marshal(version.GetVer())
			xerror.Panic(err)
			xerror.PanicErr(writer.Write(dt))
		})

		mux.Get("/dep", func(writer http.ResponseWriter, request *http.Request) {
			writer.Header().Set("Content-Type", "application/json")

			info, ok := debug.ReadBuildInfo()
			if !ok {
				writer.WriteHeader(http.StatusNoContent)
				xerror.PanicErr(writer.Write([]byte("")))
				return
			}

			dt, err := jsonx.Marshal(info)
			xerror.Panic(err)
			xerror.PanicErr(writer.Write(dt))
		})
	})
}
