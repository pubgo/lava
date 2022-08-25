package config

import (
	"github.com/pubgo/lava/internal/pkg/merge"
	"github.com/spf13/viper"
)

type CfgMap map[string]interface{}

func (t CfgMap) Decode(val interface{}) error {
	return merge.MapStruct(val, t)
}

func (t CfgMap) GetString(name string) string {
	var val, ok = t[name].(string)
	if ok {
		return val
	}
	return ""
}

type DecoderOption = viper.DecoderConfigOption
