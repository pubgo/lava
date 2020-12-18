package server

import (
	"github.com/pubgo/dix/dix_run"
	"github.com/pubgo/golug"
	"github.com/pubgo/golug/example/tickrun/server/models"
	"github.com/pubgo/golug/example/tickrun/server/router"
	"github.com/pubgo/golug/golug_db"
	"github.com/pubgo/golug/golug_entry"
	"github.com/pubgo/xerror"
)

var name = "tickrun_server"

func GetEntry() golug_entry.Entry {
	ent := golug.NewRestEntry(name, nil)
	ent.Version("v0.0.1")
	ent.Description("api server")
	ent.Router(router.Api)

	golug.WithBeforeStart(func(ctx *dix_run.BeforeStartCtx) {
		db := golug_db.GetClient()
		xerror.Exit(db.Sync2(
			new(models.Task),
		))
	})
	return ent
}
