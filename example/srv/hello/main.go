package hello

import (
	"github.com/pubgo/lava"
	"github.com/pubgo/lava/service"

	"github.com/pubgo/lava/example/pkg/proto/hellopb"
	"github.com/pubgo/lava/example/srv/hello/handler"
)

var name = "test-grpc"

func NewSrv() service.Service {
	srv := lava.NewSrv(name, "test-grpc grpc service")

	hellopb.RegisterTestApiServer(srv, handler.NewTestAPIHandler())

	return srv
}
