package config_builder

import (
	"github.com/pubgo/lava/config"
	"github.com/pubgo/lava/pkg/typex"
	"github.com/pubgo/lava/vars"
)

func init() {
	vars.Register("config_data", func() interface{} { return config.GetCfg().All() })
	vars.Register("config", func() interface{} {
		return typex.M{"cfgType": config.CfgType, "cfgName": config.CfgName, "home": config.CfgDir, "cfgPath": config.CfgPath}
	})
}
