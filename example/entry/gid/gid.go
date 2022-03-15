package gid

import (
	"github.com/pubgo/lava/example/entry/gid/handler"
	"github.com/pubgo/lava/example/protopb/proto/gid"
	"github.com/pubgo/lava/server"
	"github.com/pubgo/lava/server/grpcEntry"
)

func GetEntry() server.Entry {
	var ent = grpcEntry.New("gid")
	ent.Description("gid generate")
	// enable gateway(names...)
	// enable debug(names...)
	// enable rs(names...)
	// enable mesh(names...)
	// enable broker(names...)

	gid.RegisterIdSrvServer(ent, handler.NewId())
	return ent
}
