package config

import (
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/pubgo/funk/merge"
	"github.com/spf13/viper"
)

const (
	defaultConfigName   = "config"
	defaultConfigType   = "yaml"
	defaultConfigPath   = "./configs"
	includeConfigName   = "resources"
	componentConfigKey  = "name"
	defaultComponentKey = "default"
)

var (
	CfgDir   string
	CfgPath  string
	Replacer = strings.NewReplacer(".", "_", "-", "_", "/", "_")
)

type DecoderOption = viper.DecoderConfigOption
type CfgMap map[string]any

func (c CfgMap) Decode(val any, tags ...string) error {
	var tag = "yaml"
	if len(tags) > 0 && tags[0] != "" {
		tag = tags[0]
	}
	return merge.MapStruct(val, c, func(cfg *mapstructure.DecoderConfig) { cfg.TagName = tag }).Err()
}

type Config interface {
	UnmarshalKey(key string, rawVal interface{}, opts ...DecoderOption) error
	Unmarshal(rawVal interface{}, opts ...DecoderOption) error

	// DecodeComponent decode component config to map[string]*struct
	DecodeComponent(name string, cfgMap interface{}) error
	Get(key string) interface{}
	Set(string, interface{})
	GetString(key string) string
	AllKeys() []string
	All() map[string]interface{}
}
