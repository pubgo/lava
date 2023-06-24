package sqlite

import (
	"fmt"

	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/config"
	"github.com/pubgo/funk/errors"
	"github.com/pubgo/funk/recovery"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/pubgo/lava/core/orm"
)

func init() {
	orm.Register("postgres", func(cfg config.Node) gorm.Dialector {
		defer recovery.Raise(func(err error) error {
			return errors.WrapKV(err, "cfg", cfg)
		})

		assert.If(cfg.Get("dsn") == nil, "dsn not found")

		return postgres.New(postgres.Config{
			DSN: fmt.Sprintf("%v", cfg.Get("dsn")),
			// refer: https://github.com/go-gorm/postgres
			// disables implicit prepared statement usage. By default pgx automatically uses the extended protocol
			PreferSimpleProtocol: true,
		})
	})
}
