package bootstrap

import (
	"github.com/pubgo/funk/config"
	"github.com/pubgo/funk/recovery"
	"github.com/pubgo/lava/cmds/app"
	"github.com/pubgo/lava/internal/example/grpc/internal/handlers/gid_handler"
	"github.com/pubgo/lava/internal/example/grpc/internal/services/gid_client"
	"github.com/pubgo/lava/internal/example/grpc/taskcmd"
	"github.com/pubgo/lava/pkg/gateway"
)

func Main() {
	defer recovery.Exit()

	di := app.NewBuilder()
	di.Provide(config.Load[Config])

	di.Provide(gid_handler.New)
	di.Provide(gid_handler.NewHttp)
	di.Provide(gid_handler.NewHttp111)
	di.Provide(gid_client.New)
	di.Provide(func() *gateway.Mux {
		return gateway.NewMux()
	})
	di.Provide(taskcmd.New)
	//di.Provide(func() lava.Middleware {
	//	return lava.MiddlewareWrap{
	//		Name: "t",
	//		Next: func(next lava.HandlerFunc) lava.HandlerFunc {
	//			return func(ctx context.Context, req lava.Request) (lava.Response, error) {
	//				fmt.Println("endpoint", req.Endpoint())
	//				fmt.Println("header", req.Header().String())
	//				return next(ctx, req)
	//			}
	//		},
	//	}
	//})

	// proxy
	di.Provide(gid_handler.NewIdProxy)

	app.Run(di)
}

func MainProxy() {
	defer recovery.Exit()

	di := app.NewBuilder()
	di.Provide(config.Load[Config])

	// proxy
	di.Provide(gid_handler.NewIdProxyServer)

	app.Run(di)
}
