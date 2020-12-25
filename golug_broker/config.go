package golug_broker

import (
	"github.com/pubgo/golug/golug_consts"
)

var Name = "broker"
var cfg = make(map[string]ClientCfg)

type ClientCfg struct {
	Driver string `json:"driver"`
	Name   string `json:"name"`
}

func GetCfg() map[string]ClientCfg {
	return cfg
}

func GetDefaultCfg() ClientCfg {
	return ClientCfg{
		Driver: "nsq",
		Name:   golug_consts.Default,
	}
}
