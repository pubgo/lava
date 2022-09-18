package bootstrap

import (
	"github.com/pubgo/dix/di"
	"github.com/pubgo/lava/clients/orm"
	"github.com/pubgo/lava/config"

	"github.com/pubgo/lava/example/internal/handlers/gidrpc"
	"github.com/pubgo/lava/example/internal/services/casbinservice"
	"github.com/pubgo/lava/example/internal/services/menuservice"
)

func Init() {
	di.Provide(func() *config.App {
		return &config.App{
			Project: "hello",
		}
	})

	di.Provide(func(c config.Config) Config {
		return config.Decode[Config](c)
	})

	di.Provide(gidrpc.New)
	di.Provide(menuservice.New)
	di.Provide(casbinservice.New)
	di.Provide(orm.New)
}
