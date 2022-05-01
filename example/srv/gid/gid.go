package gid

import (
	"github.com/pubgo/lava"
	"github.com/pubgo/lava/example/protopb/proto/gid"
	"github.com/pubgo/lava/example/srv/gid/handler"
	"github.com/pubgo/lava/service"
)

func NewSrv() service.Service {
	var srv = lava.NewService("gid", "gid generate")
	gid.RegisterId(srv, handler.NewId())
	return srv
}
