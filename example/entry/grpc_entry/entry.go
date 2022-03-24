package grpc_entry

import (
	"github.com/gofiber/fiber/v2"
	"github.com/pubgo/lava/core/healthy"
	"github.com/pubgo/lava/core/registry/registry_plugin"
	"github.com/pubgo/lava/service/gateway"

	"github.com/pubgo/lava"
	"github.com/pubgo/lava/debug/debug_plugin"
	"github.com/pubgo/lava/example/entry/grpc_entry/handler"
	"github.com/pubgo/lava/example/protopb/proto/hello"
	"github.com/pubgo/lava/service/service_type"
)

var name = "test-grpc"

func GetEntry() service_type.Service {
	srv := lava.NewService(name, "entry grpc test")

	registry_plugin.Enable(srv)
	debug_plugin.Enable(srv)
	gateway.Enable(srv)

	hello.RegisterTestApi(srv, handler.NewTestAPIHandler())

	// 健康检查
	healthy.Register(name, func(req *fiber.Ctx) error {
		return nil
	})

	return srv
}
