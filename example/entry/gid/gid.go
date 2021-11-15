package gid

import (
	"github.com/pubgo/lava/entry"
	"github.com/pubgo/lava/entry/grpcEntry"
	"github.com/pubgo/lava/example/entry/gid/handler"
)

func GetEntry() entry.Entry {
	var ent = grpcEntry.New("gid")
	ent.Description("gid generate")
	ent.Register(handler.NewId())
	return ent
}