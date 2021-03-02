package grpc_entry

import (
	"github.com/pubgo/golug"
	"github.com/pubgo/golug/entry"
	"github.com/pubgo/golug/example/grpc_entry/handler"
)

var name = "test-grpc"

func GetEntry() entry.Entry {
	ent := golug.NewGrpc(name)
	ent.Version("v0.0.1")
	ent.Description("entry grpc test")
	ent.Register(handler.NewTestAPIHandler())
	return ent
}
