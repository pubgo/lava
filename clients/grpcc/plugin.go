package grpcc

import (
	"github.com/pubgo/x/merge"
	"github.com/pubgo/xerror"

	"github.com/pubgo/lava/config"
	"github.com/pubgo/lava/plugin"

	// 默认加载mdns注册中心
	_ "github.com/pubgo/lava/plugins/registry/mdns"

	// grpc log插件加载
	_ "github.com/pubgo/lava/internal/plugins/grpclog"
)

func init() {
	plugin.Register(&plugin.Base{
		Name: Name,
		OnInit: func(p plugin.Process) {
			_ = config.Decode(Name, &cfgMap)
			for name := range cfgMap {
				var defCfg = DefaultCfg()
				xerror.Panic(merge.Copy(&defCfg, cfgMap[name]))
				cfgMap[name] = defCfg
			}
		},
	})
}
