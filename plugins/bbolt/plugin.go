package bbolt

import (
	"github.com/pubgo/xerror"

	"github.com/pubgo/lug/config"
	"github.com/pubgo/lug/internal/resource"
	"github.com/pubgo/lug/plugin"
)

func init() {
	plugin.Register(&plugin.Base{
		Name: Name,
		OnInit: func(ent plugin.Entry) {
			xerror.Assert(!config.Decode(Name, &cfgMap), "config [%s] not found", Name)

			for k, v := range cfgMap {
				resource.Update(Name, k, &Client{db: v.Build()})
			}
		},
	})
}
