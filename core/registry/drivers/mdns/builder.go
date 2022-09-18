package mdns

import (
	"github.com/pubgo/dix/di"
	"github.com/pubgo/lava/core/registry"
	"github.com/pubgo/lava/logging"
	"github.com/pubgo/lava/pkg/merge"
)

func init() {
	di.Provide(func(conf *registry.Cfg, log *logging.Logger) map[string]registry.Registry {
		if conf.Driver != Name {
			return nil
		}

		var cfg Cfg
		merge.MapStruct(&cfg, conf.DriverCfg).Unwrap()
		return map[string]registry.Registry{Name: New(cfg, log)}
	})
}
