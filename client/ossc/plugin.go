package ossc

import (
	"github.com/pubgo/lug/config"
	"github.com/pubgo/lug/plugin"
	"github.com/pubgo/x/merge"
	"github.com/pubgo/xerror"
)

func init() {
	plugin.Register(&plugin.Base{
		Name: Name,
		OnInit: func(ent interface{}) {
			if !config.Decode(Name, &cfgList) {
				return
			}

			for k, v := range cfgList {
				cfg := GetDefaultCfg()
				xerror.Panic(merge.Copy(&cfg, &v))
				initClient(k, cfg)
				cfgList[k] = cfg
			}
		},
		OnVars: func(w func(name string, data func() interface{})) {
			w(Name, func() interface{} { return cfgList })
		},
	})
}
