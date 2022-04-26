package config

import (
	"github.com/pubgo/lava/module"
	"go.uber.org/fx"
)

func New() Config {
	var cfg = newCfg()
	return cfg
}

// GetCfg 获取内存配置
func GetCfg() Config {
	if conf == nil {
		panic("please init config")
	}
	return conf
}

func init() {
	module.Register("config", fx.Provide(New))
}
