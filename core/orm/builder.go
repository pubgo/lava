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

func NewClients(conf map[string]*Config, logs log.Logger) map[string]*Client {
	clients := make(map[string]*Client, len(conf))
	for name, c := range conf {
		clients[name] = New(c, logs)
	}
	return clients
}

func New(conf *Config, logs log.Logger) *Client {
	logs = logs.WithName(Name)
	conf = config.MergeR(generic.Ptr(DefaultCfg()), conf).Unwrap()

	ormCfg := merge.Copy(new(gorm.Config), conf).Unwrap()
	ormCfg.NowFunc = func() time.Time { return time.Now().UTC() }
	ormCfg.NamingStrategy = schema.NamingStrategy{TablePrefix: conf.TablePrefix}

	logCfg := DefaultLoggerCfg()
	logs.Debug().Any("config", logCfg).Msg("orm config")

	ormCfg.Logger = logger.New(log.NewStd(logs.WithCallerSkip(4)), logCfg)
	logs.Debug().Any("config", ormCfg).Msg("orm log config")

	factory := Get(conf.Driver)
	assert.If(factory == nil, "driver factory[%s] not found", conf.Driver)
	dialect := factory(conf.DriverCfg)

	db := assert.Must1(gorm.Open(dialect, ormCfg))

	// 服务连接校验
	sqlDB := assert.Must1(db.DB())
	assert.Must(sqlDB.Ping())

	if conf.MaxConnTime != 0 {
		sqlDB.SetConnMaxLifetime(conf.MaxConnTime)
	}

	if conf.MaxConnIdle != 0 {
		sqlDB.SetMaxIdleConns(conf.MaxConnIdle)
	}

	if conf.MaxConnOpen != 0 {
		sqlDB.SetMaxOpenConns(conf.MaxConnOpen)
	}

	return &Client{
		DB:          db,
		TablePrefix: conf.TablePrefix,
	}
}
