package service

import (
	"github.com/gofiber/fiber/v2"
	"github.com/urfave/cli/v2"
	"google.golang.org/grpc"

	"github.com/pubgo/lava/core/lifecycle"
	"github.com/pubgo/lava/middleware"
)

type WebHandler interface {
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

type Command interface {
	Command() *cli.Command
}

type AppInfo interface {
	Options() Options
}

type App interface {
	lifecycle.Lifecycle
	Command
	AppInfo
	Flags(flags ...cli.Flag)
	Middleware(middleware.Middleware)
	RegApp(prefix string, r *fiber.App)
}

type Service interface {
	App
	Dix(regs ...interface{})
	RegisterService(desc *grpc.ServiceDesc, impl interface{})
}
