package orm

import (
	"time"

	"github.com/pubgo/xerror"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"
	gl "gorm.io/gorm/logger"
	opentracing "gorm.io/plugin/opentracing"

	"github.com/pubgo/lava/config"
	"github.com/pubgo/lava/core/tracing"
	"github.com/pubgo/lava/logging"
	"github.com/pubgo/lava/logging/logkey"
	"github.com/pubgo/lava/pkg/merge"
	"github.com/pubgo/lava/runtime"
)

func Module() fx.Option {
	return fx.Provide(New)
}

func New(c config.Config, log *logging.Logger) map[string]*Client {
	var cfg = &Cfg{}
	xerror.Panic(c.Decode(Name, cfg))

	var ormCfg = &gorm.Config{}
	xerror.Panic(merge.Struct(ormCfg, cfg))

	var level = gl.Info
	if runtime.IsProd() || runtime.IsRelease() {
		level = gl.Error
	}

	ormCfg.Logger = gl.New(
		logPrintf(zap.L().Named(logkey.Component).Named(Name).Sugar().Infof),
		gl.Config{
			SlowThreshold:             200 * time.Millisecond,
			LogLevel:                  level,
			IgnoreRecordNotFoundError: false,
			Colorful:                  true,
		},
	)

	var factory = Get(cfg.Driver)
	xerror.Assert(factory == nil, "factory[%s] not found", cfg.Driver)
	dialect := factory(cfg.DriverCfg)

	db, err := gorm.Open(dialect, ormCfg)
	xerror.Panic(err)

	// 添加链路追踪
	xerror.Panic(db.Use(opentracing.New(
		opentracing.WithErrorTagHook(tracing.SetIfErr),
	)))

	// 服务连接校验
	sqlDB, err := db.DB()
	xerror.Panic(err)
	xerror.Panic(sqlDB.Ping())

	if cfg.MaxConnTime != 0 {
		sqlDB.SetConnMaxLifetime(cfg.MaxConnTime)
	}

	if cfg.MaxConnIdle != 0 {
		sqlDB.SetMaxIdleConns(cfg.MaxConnIdle)
	}

	if cfg.MaxConnOpen != 0 {
		sqlDB.SetMaxOpenConns(cfg.MaxConnOpen)
	}

	return nil
}
