package orm

import (
	"github.com/pubgo/xerror"

	"github.com/pubgo/lava/config"
	"github.com/pubgo/lava/pkg/merge"
	"github.com/pubgo/lava/pkg/watcher"
	"github.com/pubgo/lava/plugin"
	"github.com/pubgo/lava/resource"
	"github.com/pubgo/lava/types"
)

func init() {
	plugin.Register(&plugin.Base{
		Name: Name,
		OnInit: func(p plugin.Process) {
			xerror.Panic(config.Decode(Name, &cfgMap))
			for name := range cfgMap {
				var t = merge.Copy(DefaultCfg(), cfgMap[name]).(*Cfg)
				var factory, ok = factories.Get(cfgMap[name].Driver).(Factory)
				xerror.Assert(factory == nil || !ok, "factory[%s] not found", t.Driver)
				dialect := factory(config.GetMap(Name, name))
				var db = t.Build(dialect)
				resource.Update(name, &Client{DB: db})
				cfgMap[name] = t
			}
		},
		OnWatch: func(name string, w *watcher.Response) (err error) {
			defer xerror.RespErr(&err)
			w.OnPut(func() {
				cfg, ok := cfgMap[name]
				if !ok {
					cfg = DefaultCfg()
				}
				xerror.Panic(types.Decode(w.Value, &cfg))
				var factory = factories.Get(cfgMap[name].Driver).(Factory)
				dialect := factory(config.GetMap(Name, name))
				var db = cfg.Build(dialect)
				resource.Update(name, &Client{DB: db})
				cfgMap[name] = cfg
			})
			return nil
		},
		OnVars: func(v types.Vars) {
			v(Name+"_cfg", func() interface{} { return cfgMap })
			v(Name+"_stats", func() interface{} {
				var data = make(map[string]interface{})
				for k, v := range resource.GetByKind(Name) {
					db, err := v.(*Client).DB.DB()
					if err != nil {
						data[k] = err.Error()
					} else {
						data[k] = db.Stats()
					}
				}
				return data
			})
		},
	})
}
