package sqlite

import (
	"fmt"
	"path/filepath"

	"github.com/pubgo/lava/core/orm"

	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/errors"
	"github.com/pubgo/funk/pathutil"
	"github.com/pubgo/funk/recovery"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/pubgo/lava/core/config"
)

func init() {
	orm.Register("sqlite3", func(cfg config.Map) gorm.Dialector {
		defer recovery.Raise(func(err error) error {
			return errors.WrapKV(err, "cfg", cfg)
		})

		assert.If(cfg["dsn"] == nil, "dsn not found")

		dsn := fmt.Sprintf("%v", cfg["dsn"])
		dsn = filepath.Join(config.CfgDir, dsn)
		assert.Must(pathutil.IsNotExistMkDir(filepath.Dir(dsn)))
		return sqlite.Open(dsn)
	})
}
