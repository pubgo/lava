package driver

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"github.com/pubgo/lava/clients/orm"
	"github.com/pubgo/lava/pkg/merge"
	"github.com/pubgo/lava/types"
)

type Config struct {
	DriverName                string `json:"driver_name"`
	DSN                       string `json:"dsn"`
	SkipInitializeWithVersion bool   `json:"skip_initialize_with_version"`
	DefaultStringSize         uint   `json:"default_string_size"`
	DefaultDatetimePrecision  *int   `json:"default_datetime_precision"`
	DisableDatetimePrecision  bool   `json:"disable_datetime_precision"`
	DontSupportRenameIndex    bool   `json:"dont_support_rename_index"`
	DontSupportRenameColumn   bool   `json:"dont_support_rename_column"`
	DontSupportForShareClause bool   `json:"dont_support_for_share_clause"`
}

func init() {
	orm.Register("mysql", func(cfg types.CfgMap) (gorm.Dialector, error) {
		var conf = Config{}
		if err := cfg.Decode(&conf); err != nil {
			return nil, err
		}

		return mysql.New(*merge.Struct(&mysql.Config{}, conf).(*mysql.Config)), nil
	})
}
