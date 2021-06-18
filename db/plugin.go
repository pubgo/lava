package db

import (
	"github.com/pubgo/lug/config"
	"github.com/pubgo/lug/plugin"
	"github.com/pubgo/lug/watcher"
	"github.com/pubgo/x/merge"
	"github.com/pubgo/x/strutil"
	"github.com/pubgo/xerror"
	"xorm.io/xorm"
	"xorm.io/xorm/schemas"
)

func init() { plugin.Register(plg) }

var plg = &plugin.Base{
	Name: Name,
	OnVars: func(w func(name string, data func() interface{})) {
		w(Name+"_cfg", func() interface{} { return cfgList })
		w(Name+"_dbMetas", func() interface{} {
			var dbMetas = make(map[string][]*schemas.Table)
			xerror.Panic(clients.Each(func(key string, engine *xorm.Engine) {
				dbMetas[key] = xerror.PanicErr(engine.DBMetas()).([]*schemas.Table)
			}))
			return dbMetas
		})

		w(Name+"_sqlList", func() interface{} {
			var sqlList []string
			xerror.Panic(clients.Each(func(key string, engine *xorm.Engine) {
				var b strutil.Builder
				defer b.Reset()
				xerror.Panic(engine.DumpAll(&b))
				sqlList = append(sqlList, b.String())
			}))
			return sqlList
		})
	},
	OnInit: func(ent interface{}) {
		if !config.Decode(Name, &cfgList) {
			return
		}

		for name := range cfgList {
			cfg := GetDefaultCfg()
			xerror.Panic(merge.Copy(&cfg, cfgList[name]))

			xerror.Panic(Update(name, *cfg))
			cfgList[name] = cfg
		}
	},

	OnWatch: func(name string, w *watcher.Response) {
		cfg, ok := cfgList[name]
		if !ok {
			cfg = GetDefaultCfg()
		}

		xerror.Panic(w.Decode(&cfg))
		xerror.Panic(Update(name, *cfg))
		cfgList[name] = cfg

		w.OnDelete(func() {
			Delete(name)
		})
	},
}
