package bootstrap

import (
	"github.com/pubgo/funk/config"
	"github.com/pubgo/funk/recovery"
	"github.com/pubgo/lava/cmds/app"
	"github.com/pubgo/lava/core/orm"

	"github.com/pubgo/lava/internal/example/http/handlers/gidhandler"
	"github.com/pubgo/lava/internal/example/http/internal/migrates"
)

func Main() {
	defer recovery.Exit()
	builder := app.NewBuilder()
	builder.Provide(orm.New)
	builder.Provide(migrates.Migrations)
	builder.Provide(gidhandler.New)
	builder.Provide(config.Load[Config])
	app.Run(builder)
}
