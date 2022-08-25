package config

import (
	"github.com/pubgo/funk/assert"
)

func Decode[Cfg any]() Cfg {
	var cfg Cfg
	assert.MustF(New().Unmarshal(&cfg), "config unmarshal failed")
	return cfg
}
