package orm

import (
	"time"

	"github.com/pubgo/xerror"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	opentracing "gorm.io/plugin/opentracing"

	"github.com/pubgo/lava/logz"
	"github.com/pubgo/lava/pkg/merge"
	"github.com/pubgo/lava/plugins/tracing"
	"github.com/pubgo/lava/runenv"
)

var logs = logz.Component(Name)

var cfgMap = make(map[string]*Cfg)

type Cfg struct {
	Driver                                   string        `json:"driver" yaml:"driver"`
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

func (t Cfg) Build(dialect gorm.Dialector) *gorm.DB {
	var log = merge.Struct(&gorm.Config{}, t).(*gorm.Config)

	var level = logger.Info
	if runenv.IsProd() || runenv.IsRelease() {
		level = logger.Error
	}

	log.Logger = logger.New(
		logPrintf(logs.Infof),
		logger.Config{
			SlowThreshold:             200 * time.Millisecond,
			LogLevel:                  level,
			IgnoreRecordNotFoundError: false,
			Colorful:                  true,
		},
	)

	db, err := gorm.Open(dialect, log)
	xerror.Panic(err)

	// 添加链路追踪
	xerror.Panic(db.Use(opentracing.New(
		opentracing.WithErrorTagHook(tracing.SetIfErr),
	)))

	// 服务连接校验
	sqlDB, err := db.DB()
	xerror.Panic(err)
	xerror.Panic(sqlDB.Ping())

	if t.MaxConnTime != 0 {
		sqlDB.SetConnMaxLifetime(t.MaxConnTime)
	}

	if t.MaxConnIdle != 0 {
		sqlDB.SetMaxIdleConns(t.MaxConnIdle)
	}

	if t.MaxConnOpen != 0 {
		sqlDB.SetMaxOpenConns(t.MaxConnOpen)
	}

	return db
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
