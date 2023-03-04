package bootstrap

import (
	"github.com/pubgo/dix/di"
	"github.com/pubgo/funk/recovery"

	"github.com/pubgo/lava/internal/example/handlers/gidhandler"
	"github.com/pubgo/lava/internal/example/internal/migrates"
)

func Init() {
	defer recovery.Exit()
	di.Provide(orm.New)
	di.Provide(migrates.Migrations)
	di.Provide(config.Load[Config])
	di.Provide(gidhandler.New)
}
