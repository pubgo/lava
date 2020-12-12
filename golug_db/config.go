package golug_db

import (
	"time"

	"github.com/pubgo/golug/golug_config"
	"github.com/pubgo/xerror"
	"xorm.io/xorm/names"
)

var Name = "db"
var cfg = make(map[string]ClientCfg)

type ClientCfg struct {
	Driver      string        `json:"driver" yaml:"driver"`
	Source      string        `json:"source" yaml:"source"`
	MaxConnTime time.Duration `json:"max_conn_time" yaml:"max_conn_time"`
	MaxConnIdle int           `json:"max_conn_idle" yaml:"max_conn_idle"`
	MaxConnOpen int           `json:"max_conn_open" yaml:"max_conn_open"`
	Mapper      names.Mapper  `json:"-" yaml:"-"`
}

func GetCfg() (cfg map[string]ClientCfg) {
	defer xerror.RespExit()
	xerror.Next().Panic(golug_config.Decode(Name, &cfg))
	return
}

func GetDefaultCfg() ClientCfg {
	return ClientCfg{
		Driver:      "mysql",
		Source:      "mysql://localhost:3306/test?useUnicode=true&characterEncoding=utf-8&zeroDateTimeBehavior=convertToNull",
		MaxConnTime: time.Second * 5,
		MaxConnIdle: 10,
		MaxConnOpen: 100,
		Mapper:      names.LintGonicMapper,
	}
}
