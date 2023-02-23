package https

import (
	"github.com/pubgo/dix/di"
	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/config"
	"github.com/pubgo/funk/version"

	"github.com/pubgo/lava/pkg/fiber_builder"
)

const (
	Name               = "service"
	defaultContentType = "application/json"
)

type Config struct {
	Http       *fiber_builder.Cfg   `yaml:"http"`
	Ws         *fiber_builder.WsCfg `yaml:"ws"`
	PrintRoute bool                 `yaml:"print-route"`
	PathPrefix string               `yaml:"path-prefix"`
}

func init() {
	di.Provide(func(c config.Config) *Config {
		var cfg = Config{
			Http:       &fiber_builder.Cfg{},
			Ws:         &fiber_builder.WsCfg{},
			PrintRoute: true,
			PathPrefix: version.Project(),
		}

		assert.Must(c.UnmarshalKey(Name, &cfg))
		return &cfg
	})
}
