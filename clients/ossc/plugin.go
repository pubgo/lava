package ossc

import (
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
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
		OnInit: func(p plugin.Process) {
			xerror.Panic(config.Decode(Name, &cfgList))
			for k, v := range cfgList {
				cfg := DefaultCfg()
				xerror.Panic(merge.Copy(&cfg, &v))
				client, err := oss.New(cfg.Endpoint, cfg.AccessKeyID, cfg.AccessKeySecret)
				xerror.Panic(err)
				resource.Update(k, &Client{client})
				cfgList[k] = cfg
			}
		},
		OnWatch: func(name string, r *types.WatchResp) error {
			r.OnPut(func() {
				var cfg Cfg
				xerror.PanicF(types.Decode(r.Value, &cfg), "etcd conf parse error, cfg: %s", r.Value)

				cfg1 := DefaultCfg()
				xerror.Panic(merge.Copy(&cfg1, &cfg), "config merge error")
				client, err := oss.New(cfg.Endpoint, cfg.AccessKeyID, cfg.AccessKeySecret)
				xerror.Panic(err)
				resource.Update(name, &Client{client})
				cfgList[name] = cfg
			})

			r.OnDelete(func() {
				delete(cfgList, name)
				resource.Remove(Name, name)
			})
			return nil
		},
		OnVars: func(v types.Vars) {
			v(Name+"_cfg", func() interface{} { return cfgList })
		},
	})
}
