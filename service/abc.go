package service

import (
	"context"
	"net"

	"github.com/gofiber/fiber/v2"
	"github.com/urfave/cli/v2"
	"google.golang.org/grpc"

	"github.com/pubgo/lava/core/cmux"
	service_inter2 "github.com/pubgo/lava/internal/service_inter"
	"github.com/pubgo/lava/plugin"
)

type Desc = service_inter2.Desc
type Handler = service_inter2.Handler
type Options = service_inter2.Options
type Middleware = service_inter2.Middleware
type HandlerFunc = service_inter2.HandlerFunc
type Request = service_inter2.Request
type Response = service_inter2.Response
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
