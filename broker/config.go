package broker

import (
	"github.com/pubgo/golug/consts"
	"github.com/pubgo/golug/config"
)

var Name = "broker"

type Cfg struct {
	Driver string `json:"driver"`
	Name   string `json:"name"`
}

func GetCfg() (cfg map[string]Cfg) {
	config.Decode(Name, &cfg)
	return cfg
}

func GetDefaultCfg() Cfg {
	return Cfg{
		Driver: "nsq",
		Name:   consts.Default,
	}
}
