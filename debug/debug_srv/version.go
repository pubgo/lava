package debug_srv

import (
	"net/http"
	"os"
	rd "runtime/debug"

	"github.com/gofiber/adaptor/v2"
	"github.com/pubgo/x/jsonx"
	"github.com/pubgo/xerror"

	"github.com/pubgo/lava/debug"
	"github.com/pubgo/lava/runtime"
)

func init() {
	debug.Get("/env", adaptor.HTTPHandlerFunc(envHandle))
	debug.Get("/version", adaptor.HTTPHandlerFunc(versionHandle))
	debug.Get("/dep", adaptor.HTTPHandlerFunc(depHandle))
}

func envHandle(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Content-Type", "application/json")

	dt, err := jsonx.Marshal(os.Environ())
	xerror.Panic(err)
	xerror.PanicErr(writer.Write(dt))
}

func versionHandle(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Content-Type", "application/json")

	dt, err := jsonx.Marshal(runtime.GetVersion())
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
