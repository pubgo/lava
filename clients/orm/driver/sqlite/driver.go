package sqlite

import (
	"errors"
	"path/filepath"

	"github.com/pubgo/x/pathutil"
	"github.com/pubgo/xerror"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/pubgo/lava/clients/orm"
	"github.com/pubgo/lava/config"
	"github.com/pubgo/lava/logging/logutil"
)

func init() {
	orm.Register("sqlite3", func(cfg config.CfgMap) gorm.Dialector {
		var dsn, ok = cfg["dsn"].(string)
		xerror.AssertFn(!ok || dsn == "", func() error {
			logutil.Pretty(cfg)
			return errors.New("dsn not found")
		})
		_ = pathutil.IsNotExistMkDir(filepath.Dir(dsn))
		return sqlite.Open(dsn)
	})
}
