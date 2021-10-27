package bbolt

import (
	"github.com/pubgo/xerror"

	"github.com/pubgo/lava/config"
	"github.com/pubgo/lava/plugin"
	"github.com/pubgo/lava/resource"
)

func init() {
	plugin.Register(&plugin.Base{
		Name: Name,
		OnInit: func() {
			xerror.PanicF(config.Decode(Name, &cfgMap), "config [%s] not found", Name)

			for k, v := range cfgMap {
				var db = v.Build()
				resource.Update(k, &Client{db})
			}
		},
	})
}
