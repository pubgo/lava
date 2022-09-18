package orm

import (
	"fmt"
	"github.com/pubgo/funk/result"
	"time"

	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/recovery"
	"github.com/pubgo/funk/xerr"
	"go.uber.org/zap"
	"gorm.io/gorm"
	gl "gorm.io/gorm/logger"
	opentracing "gorm.io/plugin/opentracing"

	"github.com/pubgo/lava/core/runmode"
	"github.com/pubgo/lava/core/tracing"
	"github.com/pubgo/lava/logging"
	"github.com/pubgo/lava/logging/logutil"
	"github.com/pubgo/lava/pkg/merge"
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
	log                                      *logging.Logger
	db                                       *gorm.DB
}

func (t *Cfg) Build() (err error) {
	defer recovery.Err(&err)
	ormCfg := merge.Struct(&gorm.Config{}, t).Unwrap(func(err result.Error) result.Error {
		return err.WrapF("cfg=%#v", t)
	})

	var level = gl.Info
	if !runmode.IsDebug {
		level = gl.Error
	}

	if t.log != nil {
		ormCfg.Logger = gl.New(
			logPrintf(t.log.Named(Name).WithOptions(zap.AddCallerSkip(4)).Sugar().Infof),
			gl.Config{
				SlowThreshold:             200 * time.Millisecond,
				LogLevel:                  level,
				IgnoreRecordNotFoundError: false,
				Colorful:                  true,
			},
		)
	}

	var factory = Get(t.Driver)
	assert.If(factory == nil, "driver factory[%s] not found", t.Driver)
	dialect := factory(t.DriverCfg)

	db := assert.Must1(gorm.Open(dialect, ormCfg))

	// 添加链路追踪
	assert.Must(db.Use(opentracing.New(
		opentracing.WithErrorTagHook(tracing.SetIfErr),
	)))

	// 服务连接校验
	sqlDB := assert.Must1(db.DB())
	assert.Must(sqlDB.Ping())

	if t.MaxConnTime != 0 {
		sqlDB.SetConnMaxLifetime(t.MaxConnTime)
	}

	if t.MaxConnIdle != 0 {
		sqlDB.SetMaxIdleConns(t.MaxConnIdle)
	}

	if t.MaxConnOpen != 0 {
		sqlDB.SetMaxOpenConns(t.MaxConnOpen)
	}
	t.db = db
	return
}

func (t *Cfg) Get() *gorm.DB {
	assert.Fn(t.db == nil, func() error {
		return fmt.Errorf("please init orm")
	})
	return t.db
}

func (t *Cfg) Valid() (err error) {
	defer recovery.Err(&err, func(err xerr.XErr) xerr.XErr {
		logutil.ColorPretty(t)
		return err
	})

	assert.If(t.Driver == "", "driver is null")
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
