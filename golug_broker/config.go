package golug_broker

import (
	"github.com/pubgo/golug/golug_config"
	"github.com/pubgo/golug/golug_consts"
)

var Name = "broker"

type ClientCfg struct {
	Driver string `json:"driver"`
	Name   string `json:"name"`
}

func GetCfg() (cfg map[string]ClientCfg) {
	golug_config.Decode(Name, &cfg)
	return cfg
}

func GetDefaultCfg() ClientCfg {
	return ClientCfg{
		Driver: "nsq",
		Name:   golug_consts.Default,
	}
}
