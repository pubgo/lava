package version

import (
	"encoding/json"
	"net/http"
	"os"
	rd "runtime/debug"

	"github.com/gofiber/fiber/v3/middleware/adaptor"
	"github.com/pubgo/funk/v2/assert"

	"github.com/pubgo/lava/v2/core/debug"
	"github.com/pubgo/lava/v2/core/running"
)

func init() {
	debug.Get("env", adaptor.HTTPHandlerFunc(envHandle))
	debug.Get("version", adaptor.HTTPHandlerFunc(versionHandle))
	debug.Get("dep", adaptor.HTTPHandlerFunc(depHandle))
}

func envHandle(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Content-Type", "application/json")

	dt := assert.Must1(json.Marshal(os.Environ()))
	assert.Must1(writer.Write(dt))
}

func versionHandle(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Content-Type", "application/json")

	dt := assert.Must1(json.Marshal(running.GetSysInfo()))
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

	dt := assert.Must1(json.Marshal(info))
	assert.Must1(writer.Write(dt))
}
