package config_builder

import (
	"github.com/pubgo/xerror"
	"github.com/spf13/viper"

	"github.com/pubgo/lava/config"
	"github.com/pubgo/lava/core/watcher/watcher_builder"
)

var cfg = &configImpl{v: viper.New()}

// Init 配置初始化
func Init() {
	defer xerror.RespExit()

	cfg.initCfg()
	config.SetCfg(cfg)

	watcher_builder.Init(cfg)
}
