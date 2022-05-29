package web

import (
	"github.com/pubgo/lava"
	"github.com/pubgo/lava/example/srv/web/handlers"
	"github.com/pubgo/lava/service"
)

var name = "test-web"

func NewSrv() service.Web {
	srv := lava.NewWeb(name, "entry grpc test")
	srv.RegHandler(handlers.New())
	srv.Provide(handlers.New)
	return srv
}
