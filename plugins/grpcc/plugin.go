package grpcc

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
			if !config.Decode(Name, &configMap) {
				return
			}

			configMap.Map(func(val interface{}) interface{} {
				var cfg = val.(Cfg)
				var defCfg = GetDefaultCfg()
				xerror.Panic(merge.Copy(&defCfg, &cfg))
				return defCfg
			})
		},
	})
}
