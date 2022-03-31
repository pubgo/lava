package config

import (
	"io"

	"github.com/mitchellh/mapstructure"
)

type (
	DecoderConfig = mapstructure.DecoderConfig
	Config        interface {
		Decode(name string, fnOrPtr interface{}) error
		Get(key string) interface{}
		Set(string, interface{})
		GetString(key string) string
		GetMap(keys ...string) CfgMap
		AllKeys() []string
		MergeConfig(in io.Reader) error
		All() map[string]interface{}
	}
)
