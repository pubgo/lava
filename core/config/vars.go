package config

import "github.com/pubgo/lava/core/vars"

func init() {
	vars.Register("config", getCfgData)
	vars.Register("default-config", func() interface{} {
		return defaultCfg
	})
}
