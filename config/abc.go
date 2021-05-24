package config

import "io"

type Config interface {
	GetStringMap(key string) map[string]interface{}
	Get(key string) interface{}
	ConfigFileUsed() string
	AllKeys() []string
	MergeConfig(in io.Reader) error
}
