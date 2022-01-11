package config

import (
	"io"

	"github.com/mitchellh/mapstructure"

	"github.com/pubgo/lava/types"
)

type DecoderConfig = mapstructure.DecoderConfig
type Config interface {
	UnmarshalKey(key string, rawVal interface{}, opts ...func(*DecoderConfig)) error
	Decode(name string, fnOrPtr interface{}) error
	Get(key string) interface{}
	Set(string, interface{})
	GetString(key string) string
	GetMap(key string) types.CfgMap
	AllKeys() []string
	MergeConfig(in io.Reader) error
	All() map[string]interface{}
}
