package grpcs

import (
	"github.com/pubgo/lava/internal/pkg/fiber_builder"
	"github.com/pubgo/lava/internal/pkg/grpc_builder"

	_ "github.com/pubgo/lava/core/registry/drivers/mdns"
)

const (
	Name               = "service"
	defaultContentType = "application/grpc"
)

type Cfg struct {
	Grpc       *grpc_builder.Cfg  `yaml:"grpc-cfg"`
	Api        *fiber_builder.Cfg `yaml:"rest-cfg"`
	PrintRoute bool               `yaml:"print-route"`
}
