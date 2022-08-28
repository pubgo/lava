package bootstrap

import (
	"github.com/pubgo/dix/di"
	"github.com/pubgo/lava/clients/orm"
	"github.com/pubgo/lava/config"
	"github.com/pubgo/lava/logging"

	"github.com/pubgo/lava/example/internal/handlers/gidrpc"
	"github.com/pubgo/lava/example/internal/services/casbinservice"
	"github.com/pubgo/lava/example/internal/services/menuservice"
)

func Init() {
	di.Provide(func(c config.Config) *Config {
		return config.Decode[*Config](c)
	})

	di.Provide(gidrpc.New)

	di.Provide(func(c *Config, db *orm.Client, l *logging.Logger) *menuservice.Menu {
		return menuservice.New(c.Menu, db, l)
	})

	di.Provide(func(c *Config, db *orm.Client, l *logging.Logger) *casbinservice.Client {
		return casbinservice.New(c.Casbin, l, db)
	})

	di.Provide(func(c *Config, log *logging.Logger) *orm.Client {
		return orm.New(c.Db, log)
	})
}
