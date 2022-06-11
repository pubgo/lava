package grpcs

import (
	"github.com/pubgo/lava/internal/pkg/fiber_builder"
	"github.com/pubgo/lava/internal/pkg/grpc_builder"
)

const (
	Name               = "service"
	defaultContentType = "application/grpc"
)

type Cfg struct {
	Grpc        *grpc_builder.Cfg  `yaml:"grpc-cfg"`
	Api         *fiber_builder.Cfg `yaml:"rest-cfg"`
	Middlewares []string           `yaml:"middlewares"`
	PrintRoute  bool               `yaml:"print-route"`
}
