package version

import (
	"net/http"
	"os"
	rd "runtime/debug"

	jjson "github.com/goccy/go-json"
	"github.com/gofiber/adaptor/v2"
	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/runmode"

	"github.com/pubgo/lava/core/debug"
)

func init() {
	debug.Get("/env", adaptor.HTTPHandlerFunc(envHandle))
	debug.Get("/version", adaptor.HTTPHandlerFunc(versionHandle))
	debug.Get("/dep", adaptor.HTTPHandlerFunc(depHandle))
}

func envHandle(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Content-Type", "application/json")

	dt := assert.Must1(jjson.Marshal(os.Environ()))
	assert.Must1(writer.Write(dt))
}

func versionHandle(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Content-Type", "application/json")

	dt := assert.Must1(jjson.Marshal(runmode.GetSysInfo()))
	assert.Must1(writer.Write(dt))
}

func depHandle(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Content-Type", "application/json")

	info, ok := rd.ReadBuildInfo()
	if !ok {
		writer.WriteHeader(http.StatusNoContent)
		assert.Must1(writer.Write([]byte("")))
		return
	}

	dt := assert.Must1(jjson.Marshal(info))
	assert.Must1(writer.Write(dt))
}
