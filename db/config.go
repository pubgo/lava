package db

import (
	"github.com/pubgo/lug/config"
	"github.com/pubgo/x/pathutil"
	"github.com/pubgo/xerror"
	"xorm.io/xorm"
	xl "xorm.io/xorm/log"
	"xorm.io/xorm/names"

	"os"
	"path/filepath"
	"strings"
	"time"
)

var Name = "db"
var cfgList = make(map[string]*Cfg)

type Client struct {
	*xorm.Engine
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

	source := config.Template(cfg.Source)
	if strings.Contains(cfg.Driver, "sqlite") {
		if _dir := filepath.Dir(source); pathutil.IsNotExist(_dir) {
			_ = os.MkdirAll(_dir, 0755)
		}
	}

	engine := xerror.PanicErr(xorm.NewEngine(cfg.Driver, source)).(*xorm.Engine)
	engine.SetMaxOpenConns(cfg.MaxConnOpen)
	engine.SetMaxIdleConns(cfg.MaxConnIdle)
	engine.SetConnMaxLifetime(cfg.MaxConnTime)
	engine.SetMapper(names.LintGonicMapper)
	engine.SetLogger(newLogger("xorm"))
	engine.Logger().SetLevel(xl.LOG_DEBUG)
	engine.ShowSQL(true)
	if !cfg.Debug || config.IsStag() || config.IsProd() {
		engine.Logger().SetLevel(xl.LOG_WARNING)
		engine.ShowSQL(false)
	}

	xerror.Panic(engine.DB().Ping())
	return engine, nil
}

func GetDefaultCfg() *Cfg {
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
