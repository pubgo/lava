package hello

import (
	"github.com/pubgo/lava"
	"github.com/pubgo/lava/example/protopb/proto/hello"
	"github.com/pubgo/lava/example/srv/hello/handler"
	"github.com/pubgo/lava/service"

	_ "github.com/pubgo/lava/module/debug_module"
	_ "github.com/pubgo/lava/module/registry_module"
)

var name = "test-grpc"

func NewSrv() service.Service {
	srv := lava.NewService(name, "entry grpc test")

	hello.RegisterTestApi(srv, handler.NewTestAPIHandler())
	return srv
}
