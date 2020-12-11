package golug_log

import (
	"github.com/pubgo/golug/golug_config"
	"github.com/pubgo/xerror"
)

var Name = "log"

type Cfg map[string]ClientCfg

type ClientCfg struct {
	Driver string
	Name   string
}

func GetCfg() (cfg map[string]ClientCfg) {
	defer xerror.RespExit()
	xerror.Next().Panic(golug_config.Decode(Name, &cfg))
	return
}

func GetDefaultCfg() ClientCfg {
	return ClientCfg{

	}
}
