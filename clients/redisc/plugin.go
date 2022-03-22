package redisc

import (
	"github.com/pubgo/lava/plugin"
	"github.com/pubgo/lava/resource"
)

func init() {
	plugin.Register(&plugin.Base{
		Name: Name,
		BuilderFactory: resource.Factory{
			DefaultCfg: DefaultCfg(),
			ResType:    &Client{},
		},
	})
}
