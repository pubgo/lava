package service_type

import (
	service2 "github.com/pubgo/lava/internal/abc/service"
	"net"

	"github.com/gofiber/fiber/v2"
	"github.com/urfave/cli/v2"
	"google.golang.org/grpc"

	"github.com/pubgo/lava/core/cmux"
	"github.com/pubgo/lava/plugin"
)

type Desc = service2.Desc
type Handler = service2.Handler
type Options = service2.Options
type Middleware = service2.Middleware
type HandlerFunc = service2.HandlerFunc
type Request = service2.Request
type Response = service2.Response
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
	Middlewares() []Middleware
	RegisterApp(prefix string, r *fiber.App)
	RegisterRouter(prefix string, handlers ...fiber.Handler) fiber.Router
}
