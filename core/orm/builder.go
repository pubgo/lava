package orm

import (
	"time"

	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/config"
	"github.com/pubgo/funk/generic"
	"github.com/pubgo/funk/log"
	"github.com/pubgo/funk/merge"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

func New(conf *Config, logs log.Logger) *Client {
	logs = logs.WithName(Name)
	cfg := generic.Ptr(DefaultCfg())
	assert.Must(config.Merge(cfg, conf))
	ormCfg := merge.Copy(new(gorm.Config), cfg).Unwrap()
	ormCfg.NowFunc = func() time.Time { return time.Now().UTC() }
	ormCfg.NamingStrategy = schema.NamingStrategy{TablePrefix: cfg.TablePrefix}

	var logCfg = DefaultLoggerCfg()
	logs.Debug().Any("config", logCfg).Msg("orm config")

	ormCfg.Logger = logger.New(log.NewStd(logs.WithCallerSkip(4)), logCfg)
	logs.Debug().Any("config", ormCfg).Msg("orm log config")

	factory := Get(cfg.Driver)
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
