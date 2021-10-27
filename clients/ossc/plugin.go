package ossc

import (
	"github.com/pubgo/x/merge"
	"github.com/pubgo/xerror"

	"github.com/pubgo/lava/config"
	"github.com/pubgo/lava/plugin"
	"github.com/pubgo/lava/resource"
	"github.com/pubgo/lava/types"
)

func init() {
	plugin.Register(&plugin.Base{
		Name: Name,
		OnInit: func() {
			if config.Decode(Name, &cfgList) != nil {
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
				resource.Remove(Name, name)
			})
		},
		OnVars: func(w func(name string, data func() interface{})) {
			w(Name, func() interface{} { return cfgList })
		},
	})
}
