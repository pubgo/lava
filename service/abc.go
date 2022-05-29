package service

import (
	"github.com/gofiber/fiber/v2"
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
	Advertise string            `json:"advertise,omitempty"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

type App interface {
	Options() Options
	Command() *cli.Command
	AfterStops(...func())
	BeforeStops(...func())
	AfterStarts(...func())
	BeforeStarts(...func())
	Provide(constructors ...interface{})
	Invoke(funcs ...interface{})
	Flags(flags ...cli.Flag)
	Middleware(middleware.Middleware)
	RegApp(prefix string, r *fiber.App)
}

type Service interface {
	App
	RegisterService(desc *grpc.ServiceDesc, impl interface{})
}

type Web interface {
	App
	RegHandler(handler Handler)
}
