package service_builder

import (
	"github.com/pubgo/lava/pkg/fiber_builder"
	"github.com/pubgo/lava/pkg/grpc_builder"
	"github.com/pubgo/lava/pkg/gw_builder"
)

const (
	Name               = "service"
	defaultContentType = "application/grpc"
)

type Cfg struct {
	Grpc        grpc_builder.Cfg  `json:"grpc"`
	Api         fiber_builder.Cfg `json:"api"`
	Gw          gw_builder.Cfg    `json:"gw"`
	Advertise   string            `json:"advertise"`
	Middlewares []string          `json:"middlewares"`
}
