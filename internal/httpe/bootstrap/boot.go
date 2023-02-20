package bootstrap

import (
	"github.com/pubgo/dix/di"
	"github.com/pubgo/funk/clients/orm"
	"github.com/pubgo/funk/config"
	"github.com/pubgo/funk/recovery"

	"github.com/pubgo/lava/internal/httpe/handlers/gidhandler"
	"github.com/pubgo/lava/internal/httpe/internal/migrates"
)

func Init() {
	defer recovery.Exit()
	di.Provide(orm.New)
	di.Provide(migrates.Migrations)
	di.Provide(gidhandler.New)
	di.Provide(func() (cfg Config) {
		return config.Unmarshal(cfg)
	})
}
