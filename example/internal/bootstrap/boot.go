package bootstrap

import (
	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/result"
	"github.com/pubgo/lava/clients/orm"
	"github.com/pubgo/lava/config"
	"github.com/pubgo/lava/example/internal/handlers/gidrpc"
	"github.com/pubgo/lava/example/internal/services/casbinservice"
	"github.com/pubgo/lava/example/internal/services/menuservice"
	"github.com/pubgo/lava/logging"

	"github.com/pubgo/lava/example/internal/migrates"
)

func Providers() result.List[any] {
	return result.ListOf[any](
		gidrpc.New,
		migrates.Migrations,
		func(c config.Config) *Config {
			var cfg = new(Config)
			assert.Must(c.Unmarshal(cfg))
			return cfg
		},
		func(c *Config, db *orm.Client, l *logging.Logger) *menuservice.Menu {
			return menuservice.New(c.Menu, db, l)
		},
		func(c *Config, db *orm.Client, l *logging.Logger) *casbinservice.Client {
			return casbinservice.New(c.Casbin, l, db)
		},
		func(c *Config, log *logging.Logger) *orm.Client {
			return orm.New(c.Db, log)
		},
	)
}
