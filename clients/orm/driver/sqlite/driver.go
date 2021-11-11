package sqlite

import (
	"github.com/pubgo/x/q"
	"github.com/pubgo/xerror"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/pubgo/lava/clients/orm"
	"github.com/pubgo/lava/types"
)

func init() {
	orm.Register("sqlite3", func(cfg types.CfgMap) gorm.Dialector {
		var dsn, ok = cfg["dsn"].(string)
		xerror.AssertFn(!ok || dsn == "", func() string {
			q.Q(cfg)
			return "dns not found"
		})
		return sqlite.Open(dsn)
	})
}
