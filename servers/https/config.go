package https

import (
	"github.com/pubgo/dix/di"
	"github.com/pubgo/funk/assert"

	"github.com/pubgo/lava/pkg/fiber_builder"
)

const (
	Name               = "service"
	defaultContentType = "application/json"
)

type Cfg struct {
	Http       *fiber_builder.Cfg   `yaml:"http"`
	Ws         *fiber_builder.WsCfg `yaml:"ws"`
	PrintRoute bool                 `yaml:"print-route"`
	BasePrefix string               `yaml:"base-prefix"`
}

func init() {
	di.Provide(func(c config.Config) *Cfg {
		var cfg = Cfg{
			Http:       &fiber_builder.Cfg{},
			Ws:         &fiber_builder.WsCfg{},
			PrintRoute: true,
			BasePrefix: version.Project(),
		}

		assert.Must(c.UnmarshalKey(Name, &cfg))
		return &cfg
	})
}
