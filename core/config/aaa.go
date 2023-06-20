package config

import (
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/pubgo/funk/merge"
	"github.com/spf13/viper"
)

const (
	defaultConfigName = "config"
	defaultConfigType = "yaml"
	defaultConfigPath = "./configs"
	includeConfigName = "resources"
)

var (
	CfgDir   string
	CfgPath  string
	replacer = strings.NewReplacer(".", "_", "-", "_", "/", "_")
)

type (
	DecoderOption = viper.DecoderConfigOption
	Map           map[string]any
)

func (c Map) Decode(val any, tags ...string) error {
	tag := "yaml"
	if len(tags) > 0 && tags[0] != "" {
		tag = tags[0]
	}
	return merge.MapStruct(val, c, func(cfg *mapstructure.DecoderConfig) { cfg.TagName = tag }).Err()
}

type NamedConfig interface {
	// ConfigUniqueName unique name
	ConfigUniqueName() string
}

type Config interface {
	UnmarshalKey(key string, rawVal interface{}, opts ...DecoderOption) error
	Unmarshal(rawVal interface{}, opts ...DecoderOption) error
	Get(key string) interface{}
	Set(string, interface{})
	GetString(key string) string
	AllKeys() []string
	All() map[string]interface{}
}
