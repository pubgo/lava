package golug_codec

import (
	"github.com/pubgo/golug/golug_config"
)

var Name = "codec"

type Cfg struct {
}

func GetCfg() (cfg map[string]Cfg) {
	golug_config.Decode(Name, &cfg)
	return
}

func GetDefaultCfg() Cfg {
	return Cfg{
	}
}
