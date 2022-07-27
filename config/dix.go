package config

import (
	"github.com/pubgo/dix"
	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/recovery"
)

func init() {
	defer recovery.Exit()
	dix.Provider(func() Config { return newCfg() })
	dix.Provider(func(c Config) *App {
		var cfg App
		assert.Must(c.UnmarshalKey("app", &cfg))
		assert.Must(cfg.Check())
		return &cfg
	})
}
