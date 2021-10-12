package entry

import (
	"github.com/pubgo/lug/entry"
	"github.com/pubgo/lug/entry/grpcEntry"
	"github.com/pubgo/lug/example/gid/handler"
)

func Gid() entry.Entry {
	var ent = grpcEntry.New("gid")
	ent.Description("gid generate")
	ent.Register(handler.NewId())
	return ent
}
