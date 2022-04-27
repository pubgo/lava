package config

import (
	"github.com/pubgo/lava/module"
	"go.uber.org/fx"
)

func New() Config {
	return newCfg()
}

// GetCfg 获取内存配置
func GetCfg() Config {
	if conf == nil {
		panic("please init config")
	}
	return conf
}

func Decode(name string, cfgMap interface{}) error {
	return GetCfg().Decode(name, cfgMap)
}

func init() {
	conf = newCfg()
	module.Register(fx.Provide(New))
}
