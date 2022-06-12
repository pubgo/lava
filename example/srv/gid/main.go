package gid

import (
	"github.com/pubgo/lava"
	"github.com/pubgo/lava/service"

	"github.com/pubgo/lava/example/pkg/proto/gidpb"
	"github.com/pubgo/lava/example/srv/gid/handler"
)

func NewSrv() service.Command {
	var srv = lava.NewSrv("gid", "gid generate grpc service")
	gidpb.RegisterIdServer(srv, handler.NewId())
	return srv
}
