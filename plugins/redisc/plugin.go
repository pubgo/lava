package redisc

import (
	"github.com/pubgo/x/merge"
	"github.com/pubgo/xerror"

	"github.com/pubgo/lava/config"
	"github.com/pubgo/lava/plugin"
)

func init() {
	plugin.Register(&plugin.Base{
		Name: Name,
		OnInit: func(ent plugin.Entry) {
			_ = config.Decode(Name, &cfgMap)
			for k, v := range cfgMap {
				cfg1 := DefaultCfg()
				xerror.Panic(merge.Copy(&cfg1, v))
				Update(k, cfg1)
				cfgMap[k] = cfg1
			}
		},
	})
}
