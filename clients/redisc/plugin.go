package redisc

import (
	"github.com/pubgo/lava/plugin"
	"github.com/pubgo/lava/resource"
)

func init() {
	plugin.Register(&plugin.Base{
		Name: Name,
		Builder: resource.Base{
			OnCfg:     DefaultCfg(),
			OnWrapper: func(res resource.Resource) resource.Resource { return &Client{res} },
		},
	})
}
