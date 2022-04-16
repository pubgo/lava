package gateway_plugin

import (
	"strings"

	"github.com/gofiber/adaptor/v2"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/pubgo/xerror"

	"github.com/pubgo/lava/config"
	"github.com/pubgo/lava/pkg/gw_builder"
	"github.com/pubgo/lava/plugin"
	"github.com/pubgo/lava/service"
)

const Name = "gateway"

func Enable(srv service.Service, prefix ...string) {
	srv.Plugin(&plugin.Base{
		Name:        Name,
		CfgNotCheck: true,
		OnInit: func(p plugin.Process) {
			var cfg = gw_builder.DefaultCfg()
			if d := config.GetMap(Name); d != nil {
				xerror.Panic(d.Decode(cfg))
			}

			var builder = gw_builder.New()
			xerror.Panic(builder.Build(cfg))
			for _, desc := range srv.ServiceDesc() {
				if h, ok := desc.GrpcGatewayFn.(func(mux *runtime.ServeMux) error); ok {
					xerror.Panic(h(builder.Get()))
				}
			}

			var path = "/api"
			if len(prefix) > 0 {
				path = "/" + strings.Trim(prefix[0], "/")
			}

			srv.RegisterRouter(path).All("/*", adaptor.HTTPHandler(builder.Get()))
		},
	})
}
