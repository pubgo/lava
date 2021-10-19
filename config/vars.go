package config

import (
	"github.com/pubgo/dix"
	"github.com/pubgo/lava/vars"
)

func init() {
	vars.Watch("dix", func() interface{} { return dix.Json() })
	vars.Watch("dix_counter", func() interface{} { return float64(len(dix.Json())) })
	vars.Watch("config", func() interface{} { return GetCfg().AllSettings() })
}
