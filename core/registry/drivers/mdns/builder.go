package mdns

import (
	"github.com/pubgo/dix/di"
	"github.com/pubgo/funk/assert"
	_ "github.com/pubgo/funk/typex"
	"github.com/pubgo/lava/core/registry"
	"github.com/pubgo/lava/internal/pkg/merge"
	"github.com/pubgo/lava/logging"
)

func init() {
	di.Provide(func(conf *registry.Cfg, log *logging.Logger) map[string]registry.Registry {
		if conf.Driver != Name {
			return nil
		}

		var cfg Cfg
		assert.Must(merge.MapStruct(&cfg, conf.DriverCfg))
		return map[string]registry.Registry{Name: New(cfg, log)}
	})
}
