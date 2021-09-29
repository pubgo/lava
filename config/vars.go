package config

import (
	"github.com/pubgo/dix"
	"github.com/pubgo/lug/vars"
)

func init() {
	vars.Watch("dix", func() interface{} { return dix.Json() })
	vars.Watch("config", func() interface{} { return GetCfg().AllSettings() })
}
