package etcdv3

import (
	"github.com/pubgo/xerror"

	"github.com/pubgo/lava/config"
	"github.com/pubgo/lava/plugin"
	"github.com/pubgo/lava/resource"
	"github.com/pubgo/lava/types"
)

func init() {
	plugin.Register(&plugin.Base{
		Name: Name,
		OnInit: func(p plugin.Process) {
			xerror.Panic(config.Decode(Name, &cfgList))
			for name, cfg := range cfgList {
				etcdCfg := cfgMerge(cfg)
				client := etcdCfg.Build()
				resource.Update(name, &Client{client})
			}
		},
		OnWatch: func(name string, r *types.WatchResp) error {
			r.OnPut(func() {
				var cfg Cfg
				xerror.PanicF(types.Decode(r.Value, &cfg), "etcd conf parse error, cfg: %s", r.Value)
				etcdCfg := cfgMerge(cfg)
				client := etcdCfg.Build()
				resource.Update(name, &Client{client})
			})

			r.OnDelete(func() {
				resource.Remove(Name, name)
			})
			return nil
		},
	})
}
