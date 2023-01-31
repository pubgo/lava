package sqlite

import (
	"path/filepath"

	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/errors"
	"github.com/pubgo/funk/pathutil"
	"github.com/pubgo/funk/recovery"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/pubgo/lava/clients/orm"
	"github.com/pubgo/lava/config"
)

func init() {
	orm.Register("sqlite3", func(cfg config.CfgMap) gorm.Dialector {
		defer recovery.Raise(func(err errors.XError) {
			err.AddTag("cfg", cfg)
		})

		var dsn = cfg.GetString("dsn")
		assert.If(dsn == "", "dsn not found")
		assert.Must(pathutil.IsNotExistMkDir(filepath.Dir(dsn)))
		return sqlite.Open(dsn)
	})
}
