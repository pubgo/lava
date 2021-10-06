package ossc

import (
	"github.com/pubgo/x/merge"
	"github.com/pubgo/xerror"

	"github.com/pubgo/lug/config"
	"github.com/pubgo/lug/logger"
	"github.com/pubgo/lug/plugin"
	"github.com/pubgo/lug/types"
)

func init() { plugin.Register(plg) }

var plg = &plugin.Base{
	Name: Name,
	OnInit: func(ent plugin.Entry) {
		if !config.Decode(Name, &cfgList) {
			return
		}

		for k, v := range cfgList {
			cfg := GetDefaultCfg()
			xerror.Panic(merge.Copy(&cfg, &v))
			xerror.Panic(Update(k, cfg))
			cfgList[k] = cfg
		}
	},
	OnWatch: func(name string, r *types.WatchResp) {
		r.OnPut(func() {
			// 解析etcd配置
			var cfg ClientCfg
			xerror.PanicF(types.Decode(r.Value, &cfg), "etcd conf parse error, cfg: %s", r.Value)

			cfg1 := GetDefaultCfg()
			xerror.Panic(merge.Copy(&cfg1, &cfg), "config merge error")
			xerror.PanicF(Update(name, cfg1), "client %s watcher update error", name)
		})

		r.OnDelete(func() {
			logs.Debug("delete client", logger.Name(name))
			clientM.Delete(name)
		})
	},
	OnVars: func(w func(name string, data func() interface{})) {
		w(Name, func() interface{} { return cfgList })
	},
}
