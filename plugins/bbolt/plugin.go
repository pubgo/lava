package bbolt

import (
	"github.com/pubgo/dix"
	"github.com/pubgo/lug/config"
	"github.com/pubgo/lug/consts"
	"github.com/pubgo/lug/entry"
	"github.com/pubgo/lug/plugin"
	"github.com/pubgo/xerror"
)

func init() { plugin.Register(plg) }

var plg = &plugin.Base{
	Name: Name,
	OnInit: func(ent entry.Entry) {
		xerror.Assert(!config.Decode(Name, &cfgMap), "config [%s] not found", Name)

		var dbs = make(map[string]*DB)
		for k, v := range cfgMap {
			xerror.Exit(v.Build())
			dbs[k] = v.db
		}

		// 依赖注入DB对象
		xerror.Exit(dix.Provider(dbs))
	},
}

func Get(name ...string) *DB {
	var val, ok = cfgMap[consts.GetDefault(name...)]
	if !ok {
		return nil
	}

	return val.db
}
