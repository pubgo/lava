package etcdv3

import (
	"github.com/pubgo/lug/config"
	"github.com/pubgo/lug/consts"
	"github.com/pubgo/lug/logutil"
	"github.com/pubgo/lug/plugin"
	"github.com/pubgo/lug/watcher"

	"github.com/pubgo/xerror"
)

func init() { plugin.Register(plg) }

var plg = &plugin.Base{
	Name: Name,
	OnInit: func(ent interface{}) {
		if !config.Decode(Name, &cfgList) {
			return
		}

		for name, cfg := range cfgList {
			// etcd config处理
			cfg := xerror.PanicErr(cfgMerge(cfg)).(Cfg)
			xerror.Panic(Update(consts.GetDefault(name), cfg))
		}
	},

	OnWatch: func(name string, r *watcher.Response) {
		r.OnPut(func() {
			// 解析etcd配置
			var cfg Cfg
			xerror.PanicF(r.Decode(&cfg), "etcd conf parse error, cfg: %s", r.Value)

			cfg = xerror.PanicErr(cfgMerge(cfg)).(Cfg)
			xerror.PanicF(Update(name, cfg), "client %s watcher update error", name)
		})

		r.OnDelete(func() {
			logs.Debugf("delete client", logutil.Name(name))

			if Get(name) == nil {
				logs.Errorf("client not found", logutil.Name(name))
			}

			Delete(name)
		})
	},
}
