package hello

import (
	"github.com/pubgo/lava"
	"github.com/pubgo/lava/example/protopb/hellopb"
	"github.com/pubgo/lava/example/srv/hello/handler"
	"github.com/pubgo/lava/service"

	_ "github.com/pubgo/lava/imports/import_debug"
	_ "github.com/pubgo/lava/imports/import_registry"
)

var name = "test-grpc"

func NewSrv() service.Service {
	srv := lava.NewSrv(name, "entry grpc test")

	hellopb.RegisterTestApi(srv, handler.NewTestAPIHandler())
	return srv
}
