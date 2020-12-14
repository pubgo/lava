package golug_log

import (
	"github.com/pubgo/xlog/xlog_config"
)

var Name = "log"
var cfg = GetDefaultCfg()

func GetCfg() xlog_config.Config {
	return cfg
}

func GetDefaultCfg() xlog_config.Config {
	return xlog_config.NewProdConfig()
}
