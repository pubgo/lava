package config

import (
	"github.com/pubgo/dix"

	"github.com/pubgo/lava/pkg/typex"
	"github.com/pubgo/lava/vars"
)

func init() {
	vars.Watch("dix", func() interface{} { return dix.Json() })
	vars.Watch("config", func() interface{} { return GetCfg().All() })
	vars.Watch("config_meta", func() interface{} {
		return typex.M{"cfgType": CfgType, "cfgName": CfgName, "home": Home, "cfgPath": CfgPath}
	})
}
