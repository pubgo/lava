package etcdv3

import (
	"github.com/pubgo/xerror"

	"github.com/pubgo/lug/config"
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
				Update(name, cfg)
			}
		},
		OnWatch: func(name string, r *watcher.Response) {
			r.OnPut(func() {
				var cfg Cfg
				xerror.PanicF(types.Decode(r.Value, &cfg), "etcd conf parse error, cfg: %s", r.Value)
				Update(name, cfg)
			})

			r.OnDelete(func() {
				Delete(name)
			})
		},
	})
}
