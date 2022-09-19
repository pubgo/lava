package sqlite

import (
	"errors"

	"github.com/pubgo/funk/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/pubgo/lava/clients/orm"
	"github.com/pubgo/lava/config"
	"github.com/pubgo/lava/logging/logutil"
)

func init() {
	orm.Register("postgres", func(cfg config.CfgMap) gorm.Dialector {
		var dsn = cfg.GetString("dsn")
		assert.Fn(dsn == "", func() error {
			logutil.Pretty(cfg)
			return errors.New("dsn not found")
		})

		return postgres.New(postgres.Config{
			DSN:                  dsn,
			PreferSimpleProtocol: true,
		})
	})
}
