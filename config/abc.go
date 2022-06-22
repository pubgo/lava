package config

import (
	"io"

	"github.com/spf13/viper"

	"github.com/pubgo/lava/internal/pkg/merge"
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
type Config interface {
	LoadPath(path string)
	UnmarshalKey(key string, rawVal interface{}, opts ...DecoderOption) error
	Decode(name string, cfgMap interface{}) error
	Get(key string) interface{}
	Set(string, interface{})
	GetString(key string) string
	GetMap(keys ...string) CfgMap
	AllKeys() []string
	MergeConfig(in io.Reader) error
	All() map[string]interface{}
}
