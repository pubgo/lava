package sqlite

import (
	"github.com/pubgo/lava/config/config_type"
	"github.com/pubgo/x/pathutil"
	"github.com/pubgo/x/q"
	"github.com/pubgo/xerror"
	"path/filepath"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/pubgo/lava/clients/orm"
)

func init() {
	orm.Register("sqlite3", func(cfg config_type.CfgMap) gorm.Dialector {
		var dsn, ok = cfg["dsn"].(string)
		xerror.AssertFn(!ok || dsn == "", func() string {
			q.Q(cfg)
			return "dns not found"
		})
		_ = pathutil.IsNotExistMkDir(filepath.Dir(dsn))
		return sqlite.Open(dsn)
	})
}
