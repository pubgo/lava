package mdns

import (
	"github.com/pubgo/dix"
	"github.com/pubgo/xerror"

	"github.com/pubgo/lava/core/registry"
	"github.com/pubgo/lava/internal/pkg/merge"
	"github.com/pubgo/lava/logging"
)

func init() {
	dix.Provider(func(conf *registry.Cfg, log *logging.Logger) map[string]registry.Registry {
		if conf.Driver != Name {
			return nil
		}

		var cfg Cfg
		xerror.Panic(merge.MapStruct(&cfg, conf.DriverCfg))
		return map[string]registry.Registry{Name: New(cfg, log)}
	})
}
