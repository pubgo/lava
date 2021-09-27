package grpcc

import (
	"github.com/pubgo/x/merge"
	"github.com/pubgo/xerror"

	"github.com/pubgo/lug/config"
	"github.com/pubgo/lug/plugin"
)

func init() {
	plugin.Register(&plugin.Base{
		Name: Name,
		OnInit: func(ent plugin.Entry) {
			if !config.Decode(Name, &cfgMap) {
				return
			}

			cfgMap.Map(func(val interface{}) interface{} {
				var cfg = val.(Cfg)
				var defCfg = GetDefaultCfg()
				xerror.Panic(merge.Copy(&defCfg, &cfg))
				return defCfg
			})
		},
	})
}
