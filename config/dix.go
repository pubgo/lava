package config

import (
	"github.com/pubgo/dix"
)

func init() {
	dix.Provide(func() Config {
		return newCfg()
	})
}
