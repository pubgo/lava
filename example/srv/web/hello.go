package web

import (
	"github.com/pubgo/lava"
	"github.com/pubgo/lava/example/srv/web/handlers"
	"github.com/pubgo/lava/service"

	_ "github.com/pubgo/lava/imports/import_registry"
)

var name = "test-web"

func NewSrv() service.Web {
	srv := lava.NewWeb(name, "entry grpc test")
	srv.RegHandler(new(handlers.Handler))
	return srv
}
