package mdns

import (
	"github.com/pubgo/dix/di"
	"github.com/pubgo/funk/log"
	"github.com/pubgo/funk/merge"
	"github.com/pubgo/lava/core/registry"
)

func init() {
	di.Provide(func(conf *registry.Cfg, log log.Logger) map[string]registry.Registry {
		if conf.Driver != Name {
			return nil
		}

		var cfg Cfg
		merge.MapStruct(&cfg, conf.DriverCfg).Unwrap()
		return map[string]registry.Registry{Name: New(cfg, log)}
	})
}
