package https

import (
	"github.com/pubgo/funk/version"
	"github.com/pubgo/lava/pkg/fiberbuilder"
)

// DefaultMaxBodyBytes is the maximum allowed size of a request body in bytes.
const DefaultMaxBodyBytes = 256 * 1024

type Config struct {
	Http              *fiberbuilder.Config `yaml:"http"`
	Ws                *fiberbuilder.WsCfg  `yaml:"ws"`
	EnablePrintRouter bool                 `yaml:"enable_print_router"`
	BaseUrl           string               `yaml:"base_url"`
}

func DefaultCfg() Config {
	return Config{
		Http:              &fiberbuilder.Config{},
		Ws:                &fiberbuilder.WsCfg{},
		EnablePrintRouter: true,
		BaseUrl:           version.Project(),
	}
}
