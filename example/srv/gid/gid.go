package gid

import (
	"github.com/pubgo/lava"
	"github.com/pubgo/lava/core/debug/debug_plugin"
	"github.com/pubgo/lava/example/protopb/proto/gid"
	"github.com/pubgo/lava/example/srv/gid/handler"
	"github.com/pubgo/lava/service"
	"github.com/pubgo/lava/service/gateway_plugin"
)

func NewSrv() service.Service {
	var srv = lava.NewService("gid", "gid generate")
	gid.RegisterId(srv, handler.NewId())
	debug_plugin.Enable(srv)
	gateway_plugin.Enable(srv)
	return srv
}
