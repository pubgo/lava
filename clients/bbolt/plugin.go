package bbolt

import (
	"github.com/pubgo/lava/plugin"
	"github.com/pubgo/lava/resource"
)

func init() {
	plugin.RegisterResource(Name, resource.Factory{
		DefaultCfg: DefaultCfg(),
		ResType:    &Client{},
	})
}
