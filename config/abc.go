package config

import (
	"io"

	"github.com/spf13/viper"
)

type CfgMap = map[string]interface{}
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
