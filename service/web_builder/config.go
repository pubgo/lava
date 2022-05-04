package web_builder

import (
	"github.com/pubgo/lava/pkg/fiber_builder"
)

const (
	Name               = "web"
	defaultContentType = "application/grpc"
)

type Cfg struct {
	Api         fiber_builder.Cfg `yaml:"api"`
	Advertise   string            `yaml:"advertise"`
	Middlewares []string          `yaml:"middlewares"`
	PrintRoute  bool              `yaml:"print-route"`
}
