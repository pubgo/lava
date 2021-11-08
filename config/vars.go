package config

import (
	"github.com/pubgo/dix"

	"github.com/pubgo/lava/types"
	"github.com/pubgo/lava/vars"
)

func init() {
	vars.Register("dix", func() interface{} { return dix.Json() })
	vars.Register("config", func() interface{} { return GetCfg().All() })
	vars.Register("config_meta", func() interface{} {
		return types.M{"cfgType": CfgType, "cfgName": CfgName, "home": Home, "cfgPath": CfgPath}
	})
}
