package nsqc

import (
	"github.com/pubgo/x/merge"
	"github.com/pubgo/xerror"

	"github.com/pubgo/lava/config"
	"github.com/pubgo/lava/plugin"
	"github.com/pubgo/lava/types"
)

func init() { plugin.Register(&plg) }

var plg = plugin.Base{
	Name: Name,
	OnInit: func(p plugin.Process) {
		xerror.Panic(config.Decode(Name, &cfgList))

		for k, v := range cfgList {
			cfg := GetDefaultCfg()
			xerror.Panic(merge.Copy(&cfg, v))

			xerror.Panic(Update(k, cfg))
			cfgList[k] = cfg
		}
	},
	OnVars: func(v types.Vars) {
		v(Name, func() interface{} { return cfgList })
	},
}
