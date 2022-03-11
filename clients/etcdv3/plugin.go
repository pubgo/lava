package etcdv3

import (
	"github.com/pubgo/lava/plugin"
	"github.com/pubgo/lava/resource"
)

const Name = "etcdv3"

func init() {
	plugin.Register(&plugin.Base{
		Name: Name,
		Builder: resource.Base{
			OnCfg:     DefaultCfg(),
			OnWrapper: func(res resource.Resource) resource.Resource { return Client{res} },
		},
	})
}
