package orm

import (
	"github.com/pubgo/lava/config"
	"github.com/pubgo/lava/pkg/merge"
	"github.com/pubgo/lava/plugin"
	"github.com/pubgo/lava/resource"
)

func init() {
	plugin.Register(&plugin.Base{
		Name: Name,
		OnInit: func(p plugin.Process) {
			if config.Decode(Name, &cfgMap) != nil {
				return
			}

			for name := range cfgMap {
				cfgMap[name] = merge.Copy(DefaultCfg(), cfgMap[name]).(*Cfg)
				var db = cfgMap[name].Build()
				resource.Update(name, &Client{DB: db})
			}
		},
	})
}
