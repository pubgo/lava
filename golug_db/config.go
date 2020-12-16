package golug_db

import (
	"time"

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
	return cfg
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
