package config

import "io"

type Config interface {
	Get(key string) interface{}
	Set(string, interface{})
	GetString(key string) string
	ConfigFileUsed() string
	AllKeys() []string
	MergeConfig(in io.Reader) error
	GetStringMap(key string) map[string]interface{}
}
