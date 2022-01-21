package mdns

import (
	"github.com/pubgo/xerror"

	"github.com/pubgo/lava/plugins/registry"
	"github.com/pubgo/lava/types"
)

func init() {
	registry.Register(Name, func(cfgMap types.CfgMap) (registry.Registry, error) {
		var cfg Cfg
		xerror.Panic(cfgMap.Decode(&cfg))
		return New(cfg)
	})
}
