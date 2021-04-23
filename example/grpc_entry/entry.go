package grpc_entry

import (
	"github.com/pubgo/lug"
	"github.com/pubgo/lug/entry"
	"github.com/pubgo/lug/example/grpc_entry/handler"
)

var name = "test-grpc"

func GetEntry() entry.Entry {
	ent := lug.NewRpc(name)
	ent.Version("v0.0.1")
	ent.Description("entry grpc test")
	ent.Register(handler.NewTestAPIHandler())
	return ent
}
