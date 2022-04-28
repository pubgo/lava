package config

import (
	"io"

	"github.com/spf13/viper"

	"github.com/pubgo/lava/pkg/merge"
)

type CfgMap map[string]interface{}

func (t CfgMap) Decode(val interface{}) error {
	return merge.MapStruct(val, t)
}

type Config interface {
	UnmarshalKey(key string, rawVal interface{}, opts ...viper.DecoderConfigOption) error
	Decode(name string, cfgMap interface{}) error
	Get(key string) interface{}
	Set(string, interface{})
	GetString(key string) string
	GetMap(keys ...string) CfgMap
	AllKeys() []string
	MergeConfig(in io.Reader) error
	All() map[string]interface{}
}
