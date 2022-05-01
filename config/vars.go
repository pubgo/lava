package config

import (
	"github.com/pubgo/lava/pkg/typex"
	"github.com/pubgo/lava/vars"
)

func init() {
	vars.Register("config_data", func() interface{} {
		if conf == nil {
			return nil
		}

		return conf.All()
	})

	vars.Register("config", func() interface{} {
		return typex.M{
			"cfgType": CfgType,
			"cfgName": CfgName,
			"home":    CfgDir,
			"cfgPath": CfgPath,
		}
	})
}
