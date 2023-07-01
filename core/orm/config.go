package orm

import (
	"github.com/pubgo/funk/config"
	"time"

	"gorm.io/gorm/logger"
)

type Config struct {
	TablePrefix                              string        `yaml:"table_prefix"`
	Driver                                   string        `yaml:"driver"`
	DriverCfg                                config.Node   `yaml:"driver_config"`
	SkipDefaultTransaction                   bool          `yaml:"skip_default_transaction"`
	FullSaveAssociations                     bool          `yaml:"full_save_associations"`
	DryRun                                   bool          `yaml:"dry_run"`
	PrepareStmt                              bool          `yaml:"prepare_stmt"`
	DisableAutomaticPing                     bool          `yaml:"disable_automatic_ping"`
	DisableForeignKeyConstraintWhenMigrating bool          `yaml:"disable_foreign_key_constraint_when_migrating"`
	DisableNestedTransaction                 bool          `yaml:"disable_nested_transaction"`
	AllowGlobalUpdate                        bool          `yaml:"allow_global_update"`
	QueryFields                              bool          `yaml:"query_fields"`
	CreateBatchSize                          int           `yaml:"create_batch_size"`
	MaxConnTime                              time.Duration `yaml:"max_conn_time"`
	MaxConnIdle                              int           `yaml:"max_conn_idle"`
	MaxConnOpen                              int           `yaml:"max_conn_open"`
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
