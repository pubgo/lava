package sqlite

import (
	"path/filepath"

	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/recovery"
	"github.com/pubgo/funk/xerr"
	"github.com/pubgo/x/pathutil"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/pubgo/lava/clients/orm"
	"github.com/pubgo/lava/config"
	"github.com/pubgo/lava/logging/logutil"
)

func init() {
	defer recovery.Exit()

	orm.Register("sqlite3", func(cfg config.CfgMap) gorm.Dialector {
		defer recovery.Raise(func(err xerr.XErr) xerr.XErr {
			logutil.Pretty(cfg)
			return err
		})

		var dsn = cfg.GetString("dsn")
		assert.If(dsn == "", "dsn not found")
		_ = pathutil.IsNotExistMkDir(filepath.Dir(dsn))
		return sqlite.Open(dsn)
	})
}
