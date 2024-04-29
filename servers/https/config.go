package https

import (
	"github.com/pubgo/funk/version"
	"github.com/pubgo/lava/pkg/fiber_builder"
)

// DefaultMaxBodyBytes is the maximum allowed size of a request body in bytes.
const DefaultMaxBodyBytes = 256 * 1024

type Config struct {
	Http              *fiber_builder.Config `yaml:"http"`
	Ws                *fiber_builder.WsCfg  `yaml:"ws"`
	EnablePrintRouter bool                  `yaml:"enable_print_router"`
	BaseUrl           string                `yaml:"base_url"`
}

func DefaultCfg() Config {
	return Config{
		Http:              &fiber_builder.Config{},
		Ws:                &fiber_builder.WsCfg{},
		EnablePrintRouter: true,
		BaseUrl:           version.Project(),
	}
}
