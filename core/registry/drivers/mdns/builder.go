package mdns

import (
	"github.com/pubgo/lava/core/registry"
	"github.com/pubgo/lava/inject"
	"github.com/pubgo/lava/pkg/merge"
	"github.com/pubgo/xerror"
)

func init() {
	inject.RegGroup(registry.Name, func(conf *registry.Cfg) registry.Registry {
		if conf.Driver != Name {
			return nil
		}

		var cfg Cfg
		xerror.Panic(merge.MapStruct(&cfg, conf.DriverCfg))
		return New(cfg)
	})
}
