package bootstrap

import (
	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/result"
	"github.com/pubgo/funk/syncx"
	"github.com/pubgo/lava/clients/orm"
	"github.com/pubgo/lava/config"
	"github.com/pubgo/lava/example/internal/services/casbinservice"
	"github.com/pubgo/lava/example/internal/services/menuservice"
	"github.com/pubgo/lava/logging"

	"github.com/pubgo/lava/example/internal/migrates"
)

func Providers() result.Chan[any] {
	return syncx.Yield(func(yield func(any)) error {
		yield(migrates.Migrations)
		yield(func(c config.Config) *Config {
			var cfg = new(Config)
			assert.Must(c.Unmarshal(cfg))
			return cfg
		})

		yield(func(c *Config, db *orm.Client, l *logging.Logger) *menuservice.Menu {
			return menuservice.New(c.Menu, db, l)
		})

		yield(func(c *Config, db *orm.Client, l *logging.Logger) *casbinservice.Client {
			return casbinservice.New(c.Casbin, l, db)
		})

		yield(func(c *Config, log *logging.Logger) *orm.Client {
			return orm.New(c.Db, log)
		})
		return nil
	})
}
