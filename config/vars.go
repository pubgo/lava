package config

import (
	"github.com/pubgo/dix"
	"github.com/pubgo/lava/pkg/typex"
	"github.com/pubgo/lava/vars"
)

func init() {
	vars.Register("config_data", func() interface{} {
		var conf = new(struct {
			C Config `inject:""`
		})
		dix.Inject(&conf)
		return conf.C.All()
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
