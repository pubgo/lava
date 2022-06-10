package gid

import (
	"github.com/pubgo/lava"
	"github.com/pubgo/lava/example/pkg/proto/gidpb"
	"github.com/pubgo/lava/example/srv/gid/handler"
	"github.com/pubgo/lava/service"
)

func NewSrv() service.Command {
	var srv = lava.NewSrv("gid", "gid generate")
	srv.Dix(gidpb.RegisterIdServer)
	srv.Dix(handler.NewId)
	return srv
}
