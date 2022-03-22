package gid

import (
	"github.com/pubgo/lava"
	"github.com/pubgo/lava/example/entry/gid/handler"
	"github.com/pubgo/lava/example/protopb/proto/gid"
	"github.com/pubgo/lava/service/service_type"
)

func GetEntry() service_type.Service {
	var ent = lava.NewService("gid", "gid generate")
	// enable gateway(names...)
	// enable debug(names...)
	// enable rs(names...)
	// enable mesh(names...)
	// enable broker(names...)

	gid.RegisterId(ent, handler.NewId())
	return ent
}
