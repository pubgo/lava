package golug_log

import (
	"github.com/pubgo/golug/golug_config"
	"github.com/pubgo/xerror"
	"github.com/pubgo/xlog/xlog_config"
)

var Name = "log"
var cfg = GetDefaultCfg()

func GetCfg() (cfg xlog_config.Config) {
	defer xerror.RespExit()
	xerror.Next().Panic(golug_config.Decode(Name, &cfg))
	return
}

func GetDefaultCfg() xlog_config.Config {
	return xlog_config.NewProdConfig()
}
