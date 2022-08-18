package bootstrap

import (
	"github.com/pubgo/funk/assert"
	"github.com/pubgo/lava/clients/orm"
	"github.com/pubgo/lava/config"
	"github.com/pubgo/lava/logging"

	"github.com/pubgo/lava/example/internal/casbin"
	"github.com/pubgo/lava/example/internal/menuservice"
	"github.com/pubgo/lava/example/internal/migrates"
)

func Providers() []any {
	return []any{
		migrates.Migrations,
		func(c config.Config) *Config {
			var cfg = new(Config)
			assert.Must(c.Unmarshal(cfg))
			return cfg
		},
		func(c *Config, db *orm.Client, l *logging.Logger) *menuservice.Menu {
			return menuservice.New(c.Menu, db, l)
		},
		func(c *Config, db *orm.Client, l *logging.Logger) *casbin.Client {
			return casbin.New(c.Casbin, l, db)
		},
		func(c *Config, log *logging.Logger) *orm.Client {
			return orm.New(*c.Db, log)
		},
	}
}
