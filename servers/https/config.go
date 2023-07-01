package https

import (
	"github.com/gofiber/fiber/v2"
	"github.com/pubgo/funk/version"
	fiber_builder2 "github.com/pubgo/lava/pkg/fiber_builder"
)

type Config struct {
	Http       *fiber_builder2.Config `yaml:"http"`
	Ws         *fiber_builder2.WsCfg  `yaml:"ws"`
	PrintRoute bool                   `yaml:"print-route"`
	PathPrefix string                 `yaml:"path-prefix"`
}

func DefaultCfg() Config {
	return Config{
		Http:       &fiber_builder2.Config{},
		Ws:         &fiber_builder2.WsCfg{},
		PrintRoute: true,
		PathPrefix: version.Project(),
	}
}

func init() {
	fiber.SetParserDecoder(fiber.ParserConfig{
		IgnoreUnknownKeys: true,
		ZeroEmpty:         true,
		ParserType:        parserTypes,
	})
}
