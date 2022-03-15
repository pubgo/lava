package debug

import (
	"net/http"
	"os"
	rd "runtime/debug"

	"github.com/pubgo/x/jsonx"
	"github.com/pubgo/xerror"

	"github.com/pubgo/lava/debug/debug_mux"
	"github.com/pubgo/lava/internal/plugins/version"
)

func init() {
	debug_mux.DebugGet("/env", envHandle)
	debug_mux.DebugGet("/version", versionHandle)
	debug_mux.DebugGet("/dep", depHandle)
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
