package bootstrap

import (
	"github.com/pubgo/dix"
	"github.com/pubgo/lava/clients/orm"
	"github.com/pubgo/lava/config"
	"github.com/pubgo/lava/logging"

	"github.com/pubgo/lava/example/internal/handlers/gidrpc"
	"github.com/pubgo/lava/example/internal/services/casbinservice"
	"github.com/pubgo/lava/example/internal/services/menuservice"
)

func Init() {
	dix.Provide(func() *Config {
		return config.Decode[*Config]()
	})

	dix.Provide(gidrpc.New)

	dix.Provide(func(c *Config, db *orm.Client, l *logging.Logger) *menuservice.Menu {
		return menuservice.New(c.Menu, db, l)
	})

	dix.Provide(func(c *Config, db *orm.Client, l *logging.Logger) *casbinservice.Client {
		return casbinservice.New(c.Casbin, l, db)
	})

	dix.Provide(func(c *Config, log *logging.Logger) *orm.Client {
		return orm.New(c.Db, log)
	})
}
