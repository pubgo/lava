package versionhandler

import (
	"github.com/pubgo/lava/debug"
	"net/http"
	"os"
	rd "runtime/debug"

	"github.com/gofiber/adaptor/v2"
	"github.com/pubgo/funk/assert"
	"github.com/pubgo/x/jsonx"

	app2 "github.com/pubgo/lava/core/runmode"
)

func init() {
	debug.Get("/env", adaptor.HTTPHandlerFunc(envHandle))
	debug.Get("/version", adaptor.HTTPHandlerFunc(versionHandle))
	debug.Get("/dep", adaptor.HTTPHandlerFunc(depHandle))
}

func envHandle(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Content-Type", "application/json")

	dt := assert.Must1(jsonx.Marshal(os.Environ()))
	assert.Must1(writer.Write(dt))
}

func versionHandle(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Content-Type", "application/json")

	dt := assert.Must1(jsonx.Marshal(app2.GetVersion()))
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

	dt := assert.Must1(jsonx.Marshal(info))
	assert.Must1(writer.Write(dt))
}
