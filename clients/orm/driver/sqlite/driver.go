package sqlite

import (
	"fmt"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/pubgo/lava/clients/orm"
	"github.com/pubgo/lava/types"
)

func init() {
	orm.Register("sqlite3", func(cfg types.CfgMap) (gorm.Dialector, error) {
		var dsn, ok = cfg["dsn"].(string)
		if !ok {
			return nil, fmt.Errorf("dns not found")
		}

		return sqlite.Open(dsn), nil
	})
}
