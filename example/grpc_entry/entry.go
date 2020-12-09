package grpc_entry

import (
	"github.com/pubgo/golug"
	"github.com/pubgo/golug/example/grpc_entry/handler"
	"github.com/pubgo/golug/golug_entry"
	"github.com/pubgo/xerror"
)

func GetEntry() golug_entry.Entry {
	ent := golug.NewGrpcEntry("grpc")
	ent.Version("v0.0.1")
	ent.Description("entry grpc test")
	ent.Register(handler.NewTestAPIHandler())
	return ent
}

func GetHttpEntry() golug_entry.Entry {
	ent := golug.NewHttpEntry("grpc_api")
	ent.Version("v0.0.1")
	ent.Description("entry grpc api test")

	xerror.Panic(ent.Register(handler.NewTestAPIHandler()))
	return ent
}
