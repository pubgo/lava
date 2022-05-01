package mdns

import (
	"github.com/pubgo/lava/config"
	"github.com/pubgo/lava/core/registry"
	"github.com/pubgo/xerror"
)

func init() {
	registry.Register(Name, func(cfgMap config.CfgMap) (registry.Registry, error) {
		var cfg Cfg
		xerror.Panic(cfgMap.Decode(&cfg))
		return New(cfg)
	})
}
