package config

import (
	"github.com/pubgo/dix/di"
	"github.com/pubgo/funk/assert"
)

func init() {
	di.Provide(New)
}

func Decode[T any](c Config) T {
	assert.If(c == nil, "config is nil")

	var cfg T
	assert.Must(c.Unmarshal(&cfg))
	return cfg
}
