package orm

import (
	"time"

	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/generic"
	"github.com/pubgo/funk/log"
	"github.com/pubgo/funk/merge"
	"github.com/pubgo/funk/runmode"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

func New(cfg *Config, log log.Logger) *Client {
	log = log.WithName(Name)

	builder := merge.Struct(generic.Ptr(DefaultCfg()), cfg).Unwrap()
	ormCfg := merge.Struct(new(gorm.Config), builder).Unwrap()

	var level = logger.Info
	if !runmode.IsDebug {
		level = logger.Warn
	}

	ormCfg.NamingStrategy = schema.NamingStrategy{TablePrefix: cfg.TablePrefix}
	ormCfg.Logger = logger.New(
		log.WithName(Name).WithCallerSkip(4),
		logger.Config{
			SlowThreshold:             200 * time.Millisecond,
			LogLevel:                  level,
			IgnoreRecordNotFoundError: false,
			Colorful:                  false,
		},
	)

	var factory = Get(cfg.Driver)
	assert.If(factory == nil, "driver factory[%s] not found", cfg.Driver)
	dialect := factory(cfg.DriverCfg)

	db := assert.Must1(gorm.Open(dialect, ormCfg))

	// 服务连接校验
	sqlDB := assert.Must1(db.DB())
	assert.Must(sqlDB.Ping())

	if cfg.MaxConnTime != 0 {
		sqlDB.SetConnMaxLifetime(cfg.MaxConnTime)
	}

	if cfg.MaxConnIdle != 0 {
		sqlDB.SetMaxIdleConns(cfg.MaxConnIdle)
	}

	if cfg.MaxConnOpen != 0 {
		sqlDB.SetMaxOpenConns(cfg.MaxConnOpen)
	}

	return &Client{DB: db, TablePrefix: cfg.TablePrefix}
}
