package config

import (
	"github.com/spf13/viper"
	"go.uber.org/fx"

	"github.com/pubgo/lava/inject"
)

var conf Config

func init() {
	conf = newCfg()
	inject.Register(fx.Provide(GetCfg))
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

func GetMap(keys ...string) CfgMap {
	return GetCfg().GetMap(keys...)
}

func UnmarshalKey(key string, rawVal interface{}, opts ...viper.DecoderConfigOption) error {
	return GetCfg().UnmarshalKey(key, rawVal, opts...)
}

func GetString(key string) string {
	return GetCfg().GetString(key)
}

func Get(key string) interface{} {
	return GetCfg().Get(key)
}
