package service

import (
	"context"
	"net"

	"github.com/gofiber/fiber/v2"
	"github.com/urfave/cli/v2"
	"google.golang.org/grpc"

	"github.com/pubgo/lava/abc"
	"github.com/pubgo/lava/core/cmux"
	"github.com/pubgo/lava/pkg/typex"
)

type Desc struct {
	grpc.ServiceDesc
	Handler       interface{}
	GrpcClientFn  interface{}
	GrpcGatewayFn interface{}
}

type Handler interface {
	Close()
	Init()
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

type Service interface {
	AfterStops(...func())
	BeforeStops(...func())
	AfterStarts(...func())
	BeforeStarts(...func())
	Flags(flags ...cli.Flag)
	RegisterService(desc Desc)
	RegisterMatcher(priority int64, matches ...cmux.Matcher) chan net.Listener
	GrpcClientInnerConn() grpc.ClientConnInterface
	Plugin(name string)
	ServiceDesc() []Desc
	Options() Options
	Ctx() context.Context
	Middlewares() []abc.Middleware
	RegisterApp(prefix string, r *fiber.App)
	RegisterRouter(prefix string, fn func(r fiber.Router))
}
