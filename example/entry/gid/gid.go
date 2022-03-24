package gid

import (
	"github.com/pubgo/lava"
	"github.com/pubgo/lava/debug/debug_plugin"
	"github.com/pubgo/lava/example/entry/gid/handler"
	"github.com/pubgo/lava/example/protopb/proto/gid"
	"github.com/pubgo/lava/service/gateway"
	"github.com/pubgo/lava/service/service_type"
)

func GetEntry() service_type.Service {
	var srv = lava.NewService("gid", "gid generate")
	gid.RegisterId(srv, handler.NewId())
	debug_plugin.Enable(srv)
	gateway.Enable(srv)
	return srv
}
