package ossc

import (
	"github.com/pubgo/lava/plugin"
	"github.com/pubgo/lava/resource"
)

func init() {
	plugin.Register(&plugin.Base{
		Name: Name,
		Builder: resource.Factory{
			OnBuilder:  DefaultCfg(),
			OnResource: &Client{},
		},
	})
}
