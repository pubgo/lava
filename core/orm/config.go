package orm

import (
	"time"

	"gorm.io/gorm/logger"

	"github.com/pubgo/lava/core/config"
)

type Config struct {
	TablePrefix                              string        `json:"table_prefix" yaml:"table_prefix"`
	Driver                                   string        `json:"driver" yaml:"driver"`
	DriverCfg                                config.Node   `json:"driver_config" yaml:"driver_config"`
	SkipDefaultTransaction                   bool          `json:"skip_default_transaction" yaml:"skip_default_transaction"`
	FullSaveAssociations                     bool          `json:"full_save_associations" yaml:"full_save_associations"`
	DryRun                                   bool          `json:"dry_run" yaml:"dry_run"`
	PrepareStmt                              bool          `json:"prepare_stmt" yaml:"prepare_stmt"`
	DisableAutomaticPing                     bool          `json:"disable_automatic_ping" yaml:"disable_automatic_ping"`
	DisableForeignKeyConstraintWhenMigrating bool          `json:"disable_foreign_key_constraint_when_migrating" yaml:"disable_foreign_key_constraint_when_migrating"`
	DisableNestedTransaction                 bool          `json:"disable_nested_transaction" yaml:"disable_nested_transaction"`
	AllowGlobalUpdate                        bool          `json:"allow_global_update" yaml:"allow_global_update"`
	QueryFields                              bool          `json:"query_fields" yaml:"query_fields"`
	CreateBatchSize                          int           `json:"create_batch_size" yaml:"create_batch_size"`
	MaxConnTime                              time.Duration `json:"max_conn_time" yaml:"max_conn_time"`
	MaxConnIdle                              int           `json:"max_conn_idle" yaml:"max_conn_idle"`
	MaxConnOpen                              int           `json:"max_conn_open" yaml:"max_conn_open"`
}

func DefaultCfg() Config {
	return Config{
		SkipDefaultTransaction: true,
		MaxConnTime:            time.Hour,
		MaxConnIdle:            10,
		MaxConnOpen:            100,
	}
}

func DefaultLoggerCfg() logger.Config {
	return logger.Config{
		SlowThreshold:             200 * time.Millisecond,
		LogLevel:                  logger.Info,
		IgnoreRecordNotFoundError: false,
		Colorful:                  false,
	}
}
