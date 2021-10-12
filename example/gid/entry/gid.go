package entry

import (
	"github.com/pubgo/lava/entry"
	"github.com/pubgo/lava/entry/grpcEntry"
	"github.com/pubgo/lava/example/gid/handler"
)

func Gid() entry.Entry {
	var ent = grpcEntry.New("gid")
	ent.Description("gid generate")
	ent.Register(handler.NewId())
	return ent
}
