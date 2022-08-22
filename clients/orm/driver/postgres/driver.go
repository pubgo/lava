package sqlite

import (
	"errors"

	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/recovery"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/pubgo/lava/clients/orm"
	"github.com/pubgo/lava/config"
	"github.com/pubgo/lava/logging/logutil"
)

func init() {
	defer recovery.Exit()

	orm.Register("postgres", func(cfg config.CfgMap) gorm.Dialector {
		var dsn = cfg.GetString("dsn")
		assert.Fn(dsn == "", func() error {
			logutil.Pretty(cfg)
			return errors.New("dsn not found")
		})
		return postgres.Open(dsn)
	})
}
