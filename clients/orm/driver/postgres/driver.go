package sqlite

import (
	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/errors"
	"github.com/pubgo/funk/recovery"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/pubgo/lava/clients/orm"
	"github.com/pubgo/lava/config"
)

func init() {
	orm.Register("postgres", func(cfg config.CfgMap) gorm.Dialector {
		defer recovery.Raise(func(err errors.XError) {
			err.AddTag("cfg", cfg)
		})

		var dsn = cfg.GetString("dsn")
		assert.Fn(dsn == "", func() error {
			return errors.New("dsn not found")
		})

		return postgres.New(postgres.Config{
			DSN: dsn,
			// refer: https://github.com/go-gorm/postgres
			// disables implicit prepared statement usage. By default pgx automatically uses the extended protocol
			PreferSimpleProtocol: true,
		})
	})
}
