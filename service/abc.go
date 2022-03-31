package service_type

import (
	"context"
	"net"

	"github.com/gofiber/fiber/v2"
	"github.com/urfave/cli/v2"
	"google.golang.org/grpc"

	"github.com/pubgo/lava/core/cmux"
	"github.com/pubgo/lava/internal/abc/service_inter"
	"github.com/pubgo/lava/plugin"
)

type Desc = service_inter.Desc
type Handler = service_inter.Handler
type Options = service_inter.Options
type Middleware = service_inter.Middleware
type HandlerFunc = service_inter.HandlerFunc
type Request = service_inter.Request
type Response = service_inter.Response
type Service interface {
	AfterStops(...func())
	BeforeStops(...func())
	AfterStarts(...func())
	BeforeStarts(...func())
	Flags(flags ...cli.Flag)
	RegisterService(desc Desc)
	RegisterMatcher(priority int64, matches ...cmux.Matcher) chan net.Listener
	GrpcClientInnerConn() grpc.ClientConnInterface
	Plugin(plg plugin.Plugin)
	ServiceDesc() []Desc
	Options() Options
	Ctx() context.Context
	Middlewares() []Middleware
	RegisterApp(prefix string, r *fiber.App)
	RegisterRouter(prefix string, handlers ...fiber.Handler) fiber.Router
}
