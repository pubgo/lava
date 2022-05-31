package config

import (
	"github.com/pubgo/dix"
	"github.com/spf13/viper"
)

var conf = newCfg()

func init() {
	dix.Register(func() Config { return conf })
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

func Get(key string) string {
	return GetCfg().GetString(key)
}
