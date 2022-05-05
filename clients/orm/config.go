package orm

import (
	"time"

	"github.com/pubgo/xerror"

	"github.com/pubgo/lava/logging/logutil"
)

type Cfg struct {
	Driver                                   string                 `json:"driver" yaml:"driver"`
	DriverCfg                                map[string]interface{} `json:"driver_config" yaml:"driver_config"`
	SkipDefaultTransaction                   bool                   `json:"skip_default_transaction" yaml:"skip_default_transaction"`
	FullSaveAssociations                     bool                   `json:"full_save_associations" yaml:"full_save_associations"`
	DryRun                                   bool                   `json:"dry_run" yaml:"dry_run"`
	PrepareStmt                              bool                   `json:"prepare_stmt" yaml:"prepare_stmt"`
	DisableAutomaticPing                     bool                   `json:"disable_automatic_ping" yaml:"disable_automatic_ping"`
	DisableForeignKeyConstraintWhenMigrating bool                   `json:"disable_foreign_key_constraint_when_migrating" yaml:"disable_foreign_key_constraint_when_migrating"`
	DisableNestedTransaction                 bool                   `json:"disable_nested_transaction" yaml:"disable_nested_transaction"`
	AllowGlobalUpdate                        bool                   `json:"allow_global_update" yaml:"allow_global_update"`
	QueryFields                              bool                   `json:"query_fields" yaml:"query_fields"`
	CreateBatchSize                          int                    `json:"create_batch_size" yaml:"create_batch_size"`
	MaxConnTime                              time.Duration          `json:"max_conn_time" yaml:"max_conn_time"`
	MaxConnIdle                              int                    `json:"max_conn_idle" yaml:"max_conn_idle"`
	MaxConnOpen                              int                    `json:"max_conn_open" yaml:"max_conn_open"`
}

func (t Cfg) Valid() (err error) {
	defer xerror.Resp(func(err1 xerror.XErr) {
		err = err1
		logutil.ColorPretty(t)
	})

	xerror.Assert(t.Driver == "", "driver is null")
	return
}

func DefaultCfg() *Cfg {
	return &Cfg{
		//SkipDefaultTransaction: true,
		PrepareStmt: true,
		MaxConnTime: time.Hour,
		MaxConnIdle: 10,
		MaxConnOpen: 100,
	}
}
