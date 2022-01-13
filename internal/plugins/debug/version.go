package debug

import (
	"net/http"
	"os"
	rd "runtime/debug"

	"github.com/go-chi/chi/v5"
	"github.com/pubgo/x/jsonx"
	"github.com/pubgo/xerror"

	"github.com/pubgo/lava/internal/plugins/version"
	"github.com/pubgo/lava/mux"
)

func init() {
	mux.Debug(func(r chi.Router) {
		r.Get("/env", envHandle)
		r.Get("/version", versionHandle)
		r.Get("/dep", depHandle)
	})
}

func envHandle(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Content-Type", "application/json")

	dt, err := jsonx.Marshal(os.Environ())
	xerror.Panic(err)
	xerror.PanicErr(writer.Write(dt))
}

func versionHandle(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Content-Type", "application/json")

	dt, err := jsonx.Marshal(version.GetVer())
	xerror.Panic(err)
	xerror.PanicErr(writer.Write(dt))
}

func depHandle(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Content-Type", "application/json")

	info, ok := rd.ReadBuildInfo()
	if !ok {
		writer.WriteHeader(http.StatusNoContent)
		xerror.PanicErr(writer.Write([]byte("")))
		return
	}

	dt, err := jsonx.Marshal(info)
	xerror.Panic(err)
	xerror.PanicErr(writer.Write(dt))
}
