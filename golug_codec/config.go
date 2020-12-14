package golug_codex

import (
	"github.com/pubgo/golug/golug_config"
	"github.com/pubgo/xerror"
)

var Name = "codec"

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
