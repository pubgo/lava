package service_type

import (
	"io"
	"net"

	"github.com/gofiber/fiber/v2"
	"github.com/urfave/cli/v2"
	"google.golang.org/grpc"

	"github.com/pubgo/lava/pkg/typex"
	"github.com/pubgo/lava/plugin"
)

type Service interface {
	AfterStops(...func())
	BeforeStops(...func())
	AfterStarts(...func())
	BeforeStarts(...func())
	Flags(flags ...cli.Flag)
	RegisterService(desc Desc)
	RegisterMatcher(priority int64, matches ...func(io.Reader) bool) func() net.Listener
	GrpcClientInnerConn() grpc.ClientConnInterface
	Plugin(plugin plugin.Plugin)
	ServiceDesc() []Desc
	Options() Options
	Debug() fiber.Router
	Admin() fiber.Router
}

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
	Id       string            `json:"id,omitempty"`
	Name     string            `json:"name,omitempty"`
	Version  string            `json:"version,omitempty"`
	Port     int               `json:"port,omitempty"`
	Address  string            `json:"address,omitempty"`
	Metadata map[string]string `json:"metadata,omitempty"`
}
