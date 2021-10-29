package gin_entry

import (
	"github.com/pubgo/lava/entry"
	"github.com/pubgo/lava/entry/ginEntry"
	"github.com/pubgo/lava/internal/example/services/entry/gin_entry/handler"
)

func GetEntry() entry.Entry {
	var ent = ginEntry.New("gid1")
	ent.Description("gid1 generate")
	ent.Register(handler.NewId())
	return ent
}
