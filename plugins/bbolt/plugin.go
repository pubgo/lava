package bbolt

import (
	"github.com/pubgo/dix"
	"github.com/pubgo/xerror"

	"github.com/pubgo/lug/config"
	"github.com/pubgo/lug/plugin"
)

func init() {
	plugin.Register(&plugin.Base{
		Name: Name,
		OnInit: func(ent plugin.Entry) {
			xerror.Assert(!config.Decode(Name, &cfgMap), "config [%s] not found", Name)

			var dbs = make(map[string]*DB)
			for k, v := range cfgMap {
				xerror.Panic(v.Build())
				dbs[k] = v.db
			}

			// 依赖注入
			xerror.Panic(dix.Provider(dbs))
		},
	})
}
