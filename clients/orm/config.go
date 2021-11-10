package orm

import (
	"time"

	"github.com/pubgo/xerror"
	"gorm.io/gorm"

	"github.com/pubgo/lava/pkg/merge"
)

var cfgMap = make(map[string]*Cfg)

type Cfg struct {
	Driver                                   string        `json:"driver"`
	SkipDefaultTransaction                   bool          `json:"skip_default_transaction"`
	FullSaveAssociations                     bool          `json:"full_save_associations"`
	DryRun                                   bool          `json:"dry_run"`
	PrepareStmt                              bool          `json:"prepare_stmt"`
	DisableAutomaticPing                     bool          `json:"disable_automatic_ping"`
	DisableForeignKeyConstraintWhenMigrating bool          `json:"disable_foreign_key_constraint_when_migrating"`
	DisableNestedTransaction                 bool          `json:"disable_nested_transaction"`
	AllowGlobalUpdate                        bool          `json:"allow_global_update"`
	QueryFields                              bool          `json:"query_fields"`
	CreateBatchSize                          int           `json:"create_batch_size"`
	MaxConnTime                              time.Duration `json:"max_conn_time" yaml:"max_conn_time"`
	MaxConnIdle                              int           `json:"max_conn_idle" yaml:"max_conn_idle"`
	MaxConnOpen                              int           `json:"max_conn_open" yaml:"max_conn_open"`
}

func (t Cfg) Build() *gorm.DB {
	var dialect = factories.Get(t.Driver).(gorm.Dialector)
	xerror.Assert(dialect == nil, "dialect[%s] not found", t.Driver)
	db, err := gorm.Open(dialect, merge.Struct(&gorm.Config{}, t).(*gorm.Config))
	xerror.Panic(err)
	return db
}

func DefaultCfg() *Cfg {
	return &Cfg{
		MaxConnTime: time.Second * 5,
		MaxConnIdle: 10,
		MaxConnOpen: 100,
	}
}
