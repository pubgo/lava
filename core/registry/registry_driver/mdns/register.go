package mdns

import (
	"github.com/pubgo/xerror"
	"go.uber.org/fx"

	"github.com/pubgo/lava/core/registry"
	"github.com/pubgo/lava/inject"
	"github.com/pubgo/lava/pkg/merge"
)

func init() {
	inject.Annotated(fx.Annotated{
		Group: Name,
		Target: func(conf *registry.Cfg) registry.Registry {
			if conf.Driver == Name {
				return nil
			}

			var cfg Cfg
			xerror.Panic(merge.MapStruct(&cfg, conf.DriverCfg))
			return New(cfg)
		},
	})
}
