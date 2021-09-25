package config

import "io"

type Config interface {
	Decode(name string, fn interface{}) (b bool)
	Get(key string) interface{}
	Set(string, interface{})
	GetString(key string) string
	ConfigFileUsed() string
	AllKeys() []string
	MergeConfig(in io.Reader) error
	AllSettings() map[string]interface{}
	GetStringMap(key string) map[string]interface{}
}
