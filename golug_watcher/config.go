package golug_watcher

import (
	"github.com/pubgo/golug/golug_config"
	"github.com/pubgo/xerror"
)

type Cfg map[string]ClientCfg

type ClientCfg struct {
	Driver string `yaml:"driver"`
	Name   string `yaml:"name"`
}

func GetCfg() (cfg map[string]ClientCfg) {
	defer xerror.RespExit()
	xerror.Next().Panic(golug_config.Decode(Name, &cfg))
	return
}
