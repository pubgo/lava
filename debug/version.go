package debug

import (
	"net/http"
	"os"
	rd "runtime/debug"

	"github.com/gofiber/adaptor/v2"
	"github.com/pubgo/x/jsonx"
	"github.com/pubgo/xerror"

	app2 "github.com/pubgo/lava/core/runmode"
)

func init() {
	Get("/env", adaptor.HTTPHandlerFunc(envHandle))
	Get("/version", adaptor.HTTPHandlerFunc(versionHandle))
	Get("/dep", adaptor.HTTPHandlerFunc(depHandle))
}

func envHandle(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Content-Type", "application/json")

	dt, err := jsonx.Marshal(os.Environ())
	xerror.Panic(err)
	xerror.PanicErr(writer.Write(dt))
}

func versionHandle(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Content-Type", "application/json")

	dt, err := jsonx.Marshal(app2.GetVersion())
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
