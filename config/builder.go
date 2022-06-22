package config

import (
	"github.com/pubgo/dix"
	"github.com/pubgo/funk"
)

func init() {
	defer funk.RecoverAndExit()
	dix.Provider(func() Config { return newCfg() })
	dix.Provider(func(c Config) *App {
		var cfg App
		funk.Must(c.UnmarshalKey("app", &cfg))
		funk.Must(cfg.Check())
		return &cfg
	})
}
