package config_type

import (
	"io"

	"github.com/mitchellh/mapstructure"

	"github.com/pubgo/lava/types"
)

type (
	DecoderConfig = mapstructure.DecoderConfig
	Interface     interface {
		Decode(name string, fnOrPtr interface{}) error
		Get(key string) interface{}
		Set(string, interface{})
		GetString(key string) string
		GetMap(keys ...string) types.CfgMap
		AllKeys() []string
		MergeConfig(in io.Reader) error
		All() map[string]interface{}
	}
)
