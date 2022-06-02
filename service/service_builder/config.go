package service_builder

import (
	"github.com/pubgo/lava/pkg/fiber_builder"
	"github.com/pubgo/lava/pkg/grpc_builder"
)

const (
	Name               = "service"
	defaultContentType = "application/grpc"
)

type Cfg struct {
	Grpc        grpc_builder.Cfg  `yaml:"grpc"`
	Api         fiber_builder.Cfg `yaml:"api"`
	Advertise   string            `yaml:"advertise"`
	Middlewares []string          `yaml:"middlewares"`
	PrintRoute  bool              `yaml:"print-route"`
}
