package service

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/urfave/cli/v2"
	"google.golang.org/grpc"

	"github.com/pubgo/lava/middleware"
)

type Desc struct {
	grpc.ServiceDesc
	Handler interface{}
}

type Handler interface {
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
	Start() error
	Stop() error
	Command() *cli.Command
	Options() Options
	AfterStops(...func())
	BeforeStops(...func())
	AfterStarts(...func())
	BeforeStarts(...func())
	Provide(constructors ...interface{})
	Invoke(funcs ...interface{})
	Flags(flags ...cli.Flag)
	Middleware(middleware.Middleware)
	RegService(desc Desc)
	RegApp(prefix string, r *fiber.App)
	RegRouter(prefix string, fn func(r fiber.Router))
	RegGateway(fn func(ctx context.Context, mux *runtime.ServeMux, cc grpc.ClientConnInterface) error)
}
