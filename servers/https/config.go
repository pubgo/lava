package https

import (
	"github.com/pubgo/funk/version"
	"github.com/pubgo/lava/pkg/fiber_builder"
)

// DefaultMaxBodyBytes is the maximum allowed size of a request body in bytes.
const DefaultMaxBodyBytes = 256 * 1024

type Config struct {
	Http       *fiber_builder.Config `yaml:"http"`
	Ws         *fiber_builder.WsCfg  `yaml:"ws"`
	PrintRoute bool                  `yaml:"print-route"`
	PathPrefix string                `yaml:"path-prefix"`
}

func DefaultCfg() Config {
	return Config{
		Http:       &fiber_builder.Config{},
		Ws:         &fiber_builder.WsCfg{},
		PrintRoute: true,
		PathPrefix: version.Project(),
	}
}
