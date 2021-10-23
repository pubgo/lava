package config

import (
	"errors"
	"io"

	"github.com/pubgo/dix"
	"github.com/pubgo/xerror"

)

var ErrKeyNotFound = errors.New("config key not found")

type Config interface {
	Decode(name string, fn interface{}) error
	Get(key string) interface{}
	Set(string, interface{})
	GetString(key string) string
	ConfigPath() string
	AllKeys() []string
	MergeConfig(in io.Reader) error
	All() map[string]interface{}
	GetMap(key string) map[string]interface{}
}

func On(fn func(cfg Config)) {
	xerror.Exit(dix.Provider(fn))
}
