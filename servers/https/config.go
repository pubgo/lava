package https

import (
	"github.com/pubgo/funk/version"
	"github.com/pubgo/lava/internal/fiber_builder"
)

type Config struct {
	Http       *fiber_builder.Config `yaml:"http"`
	Ws         *fiber_builder.WsCfg  `yaml:"ws"`
	PrintRoute bool                  `yaml:"print-route"`
	PathPrefix string                `yaml:"path-prefix"`
}

func defaultCfg() Config {
	return Config{
		Http:       &fiber_builder.Config{},
		Ws:         &fiber_builder.WsCfg{},
		PrintRoute: true,
		PathPrefix: version.Project(),
	}
}
