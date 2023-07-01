package bootstrap

import (
	"github.com/pubgo/dix"
	"github.com/pubgo/funk/config"
	"github.com/pubgo/funk/recovery"
	"github.com/pubgo/lava/cmds/builder"
	"github.com/pubgo/lava/core/orm"

	"github.com/pubgo/lava/internal/example/http/handlers/gidhandler"
	"github.com/pubgo/lava/internal/example/http/internal/migrates"
)

func Main() {
	defer recovery.Exit()
	di := builder.NewDix(dix.WithValuesNull())
	di.Provide(orm.New)
	di.Provide(migrates.Migrations)
	di.Provide(gidhandler.New)
	di.Provide(config.Load[Config])
	builder.Run(di)
}
