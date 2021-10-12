package nsqc

import (
	"github.com/pubgo/x/merge"
	"github.com/pubgo/xerror"

	"github.com/pubgo/lava/config"
	"github.com/pubgo/lava/plugin"
)

func init() { plugin.Register(&plg) }

var plg = plugin.Base{
	Name: Name,
	OnInit: func(ent plugin.Entry) {
		config.Decode(Name, &cfgList)

		for k, v := range cfgList {
			cfg := GetDefaultCfg()
			xerror.Panic(merge.Copy(&cfg, v))

			xerror.Panic(Update(k, cfg))
			cfgList[k] = cfg
		}
	},
	OnVars: func(w func(name string, data func() interface{})) {
		w(Name, func() interface{} { return cfgList })
	},
}
