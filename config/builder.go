package config

import (
	"github.com/pubgo/dix"
	"github.com/pubgo/xerror"
)

func init() {
	dix.Provider(func() Config { return newCfg() })
	dix.Provider(func(c Config) *App {
		var cfg App
		xerror.Panic(c.UnmarshalKey("app", &cfg))
		xerror.Panic(cfg.Check())
		return &cfg
	})
}
