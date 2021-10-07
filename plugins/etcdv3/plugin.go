package etcdv3

import (
	"github.com/pubgo/xerror"

	"github.com/pubgo/lug/config"
	"github.com/pubgo/lug/consts"
	"github.com/pubgo/lug/logger"
	"github.com/pubgo/lug/plugin"
	"github.com/pubgo/lug/types"
	"github.com/pubgo/lug/watcher"
)

func init() {
	plugin.Register(&plugin.Base{
		Name: Name,
		OnInit: func(ent plugin.Entry) {
			_ = config.Decode(Name, &cfgList)
			for name, cfg := range cfgList {
				// etcd config处理
				cfg := xerror.PanicErr(cfgMerge(cfg)).(Cfg)
				Update(consts.GetDefault(name), cfg)
			}
		},
		OnWatch: func(name string, r *watcher.Response) {
			r.OnPut(func() {
				// 解析etcd配置
				var cfg Cfg
				xerror.PanicF(types.Decode(r.Value, &cfg), "etcd conf parse error, cfg: %s", r.Value)

				cfg = xerror.PanicErr(cfgMerge(cfg)).(Cfg)
				Update(name, cfg)
			})

			r.OnDelete(func() {
				logs.Debug("delete client", logger.Name(name))
				Delete(name)
			})
		},
	})
}
