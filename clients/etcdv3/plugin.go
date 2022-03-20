package etcdv3

import (
	"github.com/pubgo/lava/plugin"
	"github.com/pubgo/lava/resource"
)

const Name = "etcdv3"

func init() {
	plugin.Register(&plugin.Base{
		Name: Name,
		BuilderFactory: resource.Factory{
			CfgBuilder: DefaultCfg(),
			ResType:    &Client{},
		},
	})
}
