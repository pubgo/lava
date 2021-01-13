package golug_watcher

import (
	"github.com/pubgo/golug/golug_config"
)

var Name = "watcher"

type Cfg struct {
	Project string `json:"project"`
	Driver  string `json:"driver"`
	Name    string `json:"name"`
}

func GetCfg() (cfg map[string]Cfg) {
	golug_config.Decode(Name, &cfg)
	return
}

func GetDefaultCfg() Cfg {
	return Cfg{}
}
