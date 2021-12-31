package config

import (
	"errors"
	"io"

	"github.com/spf13/viper"
)

var ErrKeyNotFound = errors.New("config key not found")

type Config interface {
	UnmarshalKey(key string, rawVal interface{}, opts ...viper.DecoderConfigOption) error
	Decode(name string, fn interface{}) error
	Get(key string) interface{}
	Set(string, interface{})
	GetString(key string) string
	GetMap(key string) map[string]interface{}
	ConfigPath() string
	AllKeys() []string
	MergeConfig(in io.Reader) error
	All() map[string]interface{}
}
