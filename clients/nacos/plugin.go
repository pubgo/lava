package nacos

import (
	"github.com/nacos-group/nacos-sdk-go/clients"
	"github.com/nacos-group/nacos-sdk-go/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/vo"
	"github.com/pubgo/x/merge"
	"github.com/pubgo/xerror"

	"github.com/pubgo/lava/config"
	"github.com/pubgo/lava/plugin"
)

func init() {
	plugin.Register(&plugin.Base{
		Name: Name,
		OnInit: func(p plugin.Process) {
			_ = config.Decode(Name, &cfgMap)
			for _, cfg := range cfgMap {
				var ncp vo.NacosClientParam
				xerror.Panic(merge.CopyStruct(&ncp, cfg))
				cfg.c = &Client{
					cfg: xerror.PanicErr(clients.NewConfigClient(ncp)).(config_client.IConfigClient),
					srv: xerror.PanicErr(clients.NewNamingClient(ncp)).(naming_client.INamingClient),
				}
			}
		},
	})
}
