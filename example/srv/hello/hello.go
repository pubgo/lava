package hello

import (
	"github.com/pubgo/lava"
	"github.com/pubgo/lava/example/protopb/hellopb"
	"github.com/pubgo/lava/example/srv/hello/handler"
	"github.com/pubgo/lava/service"
)

var name = "test-grpc"

func NewSrv() service.Service {
	srv := lava.NewSrv(name, "entry grpc test")
	srv.Dix(hellopb.RegisterTestApiServer)
	srv.Dix(handler.NewTestAPIHandler)

	return srv
}
