package nsqc

import (
	"github.com/pubgo/dix"
	"github.com/pubgo/xerror"

	"github.com/pubgo/lava/config"
	"github.com/pubgo/lava/pkg/merge"
	"github.com/pubgo/lava/plugin"
	"github.com/pubgo/lava/types"
)

func init() {
	plugin.Register(&plugin.Base{
		Name: Name,
		OnInit: func(p plugin.Process) {
			xerror.Panic(config.Decode(Name, &cfgList))

			for name, v := range cfgList {
				cfg := GetDefaultCfg()
				merge.Copy(&cfg, v)

				xerror.Assert(name == "", "[name] should not be null")

				// 创建新的客户端
				client, err := cfg.Build()
				xerror.Panic(err)
				cfgList[name] = cfg
				xerror.Panic(dix.Provider(client))
			}
		},
		OnVars: func(v types.Vars) {
			v(Name, func() interface{} { return cfgList })
		},
	})
}
