package rsync

import (
	"github.com/pubgo/lava/entry/grpcEntry"
	"github.com/pubgo/lava/internal/example/services/entry/gid/handler"
)

func New() grpcEntry.Entry {
	var ent = grpcEntry.New("rsync")
	ent.Description("gid generate")
	ent.Register(handler.NewId())
	return ent
}
