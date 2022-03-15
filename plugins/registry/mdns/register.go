package mdns

import (
	"github.com/pubgo/lava/config/config_type"
	"github.com/pubgo/xerror"

	"github.com/pubgo/lava/plugins/registry"
)

func init() {
	registry.Register(Name, func(cfgMap config_type.CfgMap) (registry.Registry, error) {
		var cfg Cfg
		xerror.Panic(cfgMap.Decode(&cfg))
		return New(cfg)
	})
}
