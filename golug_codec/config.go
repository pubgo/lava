package golug_codec

import (
	"github.com/pubgo/golug/golug_config"
)

var Name = "codec"

type Cfg map[string]ClientCfg

type ClientCfg struct {
}

func GetCfg() (cfg map[string]ClientCfg) {
	golug_config.Decode(Name, &cfg)
	return
}

func GetDefaultCfg() ClientCfg {
	return ClientCfg{
	}
}
