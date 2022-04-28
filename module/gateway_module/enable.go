package gateway_module

import (
	"context"

	"github.com/gofiber/adaptor/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/pubgo/xerror"
	"go.uber.org/fx"

	"github.com/pubgo/lava/config"
	"github.com/pubgo/lava/module"
	"github.com/pubgo/lava/pkg/gw_builder"
	"github.com/pubgo/lava/service"
)

const Name = "gateway"

var path = "/api/gw"

func init() {
	module.Register(fx.Invoke(Enable))
}

func Enable(srv service.Service) {
	srv.RegisterRouter(path, func(r fiber.Router) {
		var cfg = gw_builder.DefaultCfg()
		xerror.Panic(config.UnmarshalKey(Name, &cfg))

		var builder = gw_builder.New()
		xerror.Panic(builder.Build(cfg))
		for _, desc := range srv.ServiceDesc() {
			if desc.GrpcGatewayFn == nil {
				continue
			}

			xerror.Panic(desc.GrpcGatewayFn(context.Background(), builder.Get(), srv.InnerConn()))
		}

		r.All("/*", adaptor.HTTPHandler(builder.Get()))
	})
}
