package mdns

import (
	"github.com/pubgo/lava/config/config_type"
	"github.com/pubgo/lava/core/registry"
	"github.com/pubgo/lava/core/registry/registry_type"
	"github.com/pubgo/xerror"
)

func init() {
	registry.Register(Name, func(cfgMap config_type.CfgMap) (registry_type.Registry, error) {
		var cfg Cfg
		xerror.Panic(cfgMap.Decode(&cfg))
		return New(cfg)
	})
}
