package config

import (
	"github.com/pubgo/lug/vars"
)

func init() {
	vars.Watch("config", func() interface{} { return GetCfg().AllSettings() })
}
