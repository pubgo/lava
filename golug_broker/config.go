package golug_broker

import (
	"github.com/pubgo/golug/golug_config"
	"github.com/pubgo/golug/golug_consts"
	"github.com/pubgo/xerror"
)

var Name = "broker"

type Cfg map[string]ClientCfg

type ClientCfg struct {
	Driver string `json:"driver"`
	Name   string `json:"name"`
}

func GetCfg() (cfg map[string]ClientCfg) {
	xerror.Next().Panic(golug_config.Decode(Name, &cfg))
	return
}

func GetDefaultCfg() ClientCfg {
	return ClientCfg{
		Driver: "nsq",
		Name:   golug_consts.Default,
	}
}
