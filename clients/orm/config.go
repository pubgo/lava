package orm

import (
	"github.com/pubgo/xerror"
	"gorm.io/gorm"

	"github.com/pubgo/lava/pkg/merge"
)

type Cfg struct {
	Driver                                   string `json:"driver"`
	SkipDefaultTransaction                   bool   `json:"skip_default_transaction"`
	FullSaveAssociations                     bool   `json:"full_save_associations"`
	DryRun                                   bool   `json:"dry_run"`
	PrepareStmt                              bool   `json:"prepare_stmt"`
	DisableAutomaticPing                     bool   `json:"disable_automatic_ping"`
	DisableForeignKeyConstraintWhenMigrating bool   `json:"disable_foreign_key_constraint_when_migrating"`
	DisableNestedTransaction                 bool   `json:"disable_nested_transaction"`
	AllowGlobalUpdate                        bool   `json:"allow_global_update"`
	QueryFields                              bool   `json:"query_fields"`
	CreateBatchSize                          int    `json:"create_batch_size"`
}

func (t Cfg) Builder() *gorm.DB {
	var dia = factories.Get(t.Driver).(gorm.Dialector)
	if dia == nil {

	}

	db, err := gorm.Open(dia, merge.Struct(&gorm.Config{}, t).(*gorm.Config))
	xerror.Panic(err)
	return db
}
