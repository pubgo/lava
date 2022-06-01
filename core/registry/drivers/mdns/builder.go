package mdns

import (
	"github.com/pubgo/dix"
	"github.com/pubgo/xerror"

	"github.com/pubgo/lava/core/registry"
	"github.com/pubgo/lava/pkg/merge"
)

func init() {
	dix.Register(func(conf *registry.Cfg) map[string]registry.Registry {
		if conf.Driver != Name {
			return nil
		}

		var cfg Cfg
		xerror.Panic(merge.MapStruct(&cfg, conf.DriverCfg))
		return map[string]registry.Registry{Name: New(cfg)}
	})
}
