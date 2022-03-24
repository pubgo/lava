package service_inter

import (
	"github.com/gofiber/fiber/v2"
	"google.golang.org/grpc"

	"github.com/pubgo/lava/pkg/typex"
)

type Desc struct {
	grpc.ServiceDesc
	Handler       interface{}
	GrpcClientFn  interface{}
	GrpcGatewayFn interface{}
}

type Handler interface {
	Init() func()
	Flags() typex.Flags
	Router(r fiber.Router)
}

type Options struct {
	Id        string            `json:"id,omitempty"`
	Name      string            `json:"name,omitempty"`
	Version   string            `json:"version,omitempty"`
	Port      int               `json:"port,omitempty"`
	Address   string            `json:"address,omitempty"`
	Advertise string            `json:"advertise"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}
