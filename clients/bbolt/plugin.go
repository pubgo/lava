package bbolt

import (
	"github.com/pubgo/xerror"

	"github.com/pubgo/lava/config"
	"github.com/pubgo/lava/pkg/merge"
	"github.com/pubgo/lava/plugin"
	"github.com/pubgo/lava/resource"
)

func init() {
	plugin.Register(&plugin.Base{
		Name: Name,
		OnInit: func(p plugin.Process) {
			xerror.PanicF(config.Decode(Name, &cfgMap), "config [%s] not found", Name)

			for k, v := range cfgMap {
				cfgMap[k] = merge.Struct(DefaultCfg(), v).(*Cfg)
				var db = cfgMap[k].Build()
				resource.Update(k, &Client{db})
			}
		},
	})
}
