package golug_watcher

import (
	"github.com/pubgo/golug/golug_config"
	"github.com/pubgo/xerror"
)

var Name = "watcher"

type ClientCfg struct {
	Project string
	Driver  string
	Name    string
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
