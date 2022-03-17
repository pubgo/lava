package gid

import (
	"github.com/pubgo/lava/entry"
	"github.com/pubgo/lava/entry/grpcEntry"
	"github.com/pubgo/lava/example/entry/gid/handler"
	"github.com/pubgo/lava/example/protopb/proto/gid"
)

func GetEntry() entry.Entry {
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
