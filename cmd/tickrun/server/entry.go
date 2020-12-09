package server

import (
	"github.com/pubgo/golug"
	"github.com/pubgo/golug/cmd/tickrun/server/router"
	"github.com/pubgo/golug/golug_entry"
)

var name = "tickrun_server"

func GetEntry() golug_entry.Entry {
	ent := golug.NewHttpEntry(name)
	ent.Version("v0.0.1")
	ent.Description("api server")
	ent.Router("/", router.Api)
	return ent
}
