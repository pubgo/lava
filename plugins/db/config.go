package db

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pubgo/x/pathutil"
	"github.com/pubgo/xerror"
	"go.uber.org/zap"
	"xorm.io/xorm"
	xl "xorm.io/xorm/log"
	"xorm.io/xorm/names"

	"github.com/pubgo/lug/config"
	"github.com/pubgo/lug/logger"
	"github.com/pubgo/lug/runenv"
)

var Name = "db"
var cfgList = make(map[string]*Cfg)
var logs = zap.L()

func init() {
	logs = logger.On(func(log *zap.Logger) { logs = log.Named(Name) })
}

type Cfg struct {
	Debug       bool          `json:"debug" yaml:"debug"`
	Driver      string        `json:"driver" yaml:"driver"`
	Source      string        `json:"source" yaml:"source"`
	MaxConnTime time.Duration `json:"max_conn_time" yaml:"max_conn_time"`
	MaxConnIdle int           `json:"max_conn_idle" yaml:"max_conn_idle"`
	MaxConnOpen int           `json:"max_conn_open" yaml:"max_conn_open"`
	Mapper      names.Mapper  `json:"-" yaml:"-"`
}

func (cfg Cfg) Build() (_ *xorm.Engine, err error) {
	defer xerror.RespErr(&err)

	if strings.Contains(cfg.Driver, "sqlite") {
		if !filepath.IsAbs(cfg.Source) {
			cfg.Source = filepath.Join(config.Home, cfg.Source)
		}
		if rootDir := filepath.Dir(cfg.Source); pathutil.IsNotExist(rootDir) {
			_ = os.MkdirAll(rootDir, 0755)
		}
	}

	engine := xerror.PanicErr(xorm.NewEngine(cfg.Driver, cfg.Source)).(*xorm.Engine)
	engine.SetMaxOpenConns(cfg.MaxConnOpen)
	engine.SetMaxIdleConns(cfg.MaxConnIdle)
	engine.SetConnMaxLifetime(cfg.MaxConnTime)
	engine.SetMapper(names.LintGonicMapper)
	engine.SetLogger(newLogger())
	engine.Logger().SetLevel(xl.LOG_DEBUG)
	engine.ShowSQL(true)
	if !cfg.Debug || runenv.IsStag() || runenv.IsProd() {
		engine.Logger().SetLevel(xl.LOG_WARNING)
		engine.ShowSQL(false)
	}

	return engine, xerror.Wrap(engine.DB().Ping())
}

func DefaultCfg() *Cfg {
	return &Cfg{
		Debug:       true,
		Driver:      "mysql",
		Source:      "mysql://localhost:3306/test?useUnicode=true&characterEncoding=utf-8&zeroDateTimeBehavior=convertToNull",
		MaxConnTime: time.Second * 5,
		MaxConnIdle: 10,
		MaxConnOpen: 100,
		Mapper:      names.LintGonicMapper,
	}
}
