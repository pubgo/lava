package service

import (
	"github.com/pubgo/lava/pkg/fiber_builder"
	"github.com/pubgo/lava/pkg/grpc_builder"
)

const Name = "service"

const (
	defaultContentType = "application/grpc"
)

type Cfg struct {
	Grpc      grpc_builder.Cfg  `json:"grpc"`
	Gw        fiber_builder.Cfg `json:"gw"`
	Advertise string            `json:"advertise"`

	id       string
	name     string
	hostname string
}
