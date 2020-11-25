package grpc_entry

import (
	"github.com/pubgo/golug"
	"github.com/pubgo/golug/example/grpc_entry/handler"
	"github.com/pubgo/golug/golug_entry"
	"github.com/pubgo/xerror"
)

func GetEntry() golug_entry.Entry {
	ent := golug.NewGrpcEntry("grpc")
	xerror.Panic(ent.Version("v0.0.1"))
	xerror.Panic(ent.Description("entry http test"))

	ent.Register(handler.NewTestAPIHandler())
	return ent
}

func GetHttpEntry() golug_entry.Entry {
	ent := golug.NewHttpEntry("grpc_api")
	xerror.Panic(ent.Version("v0.0.1"))
	xerror.Panic(ent.Description("entry http test"))

	xerror.Panic(ent.Register(handler.NewTestAPIHandler()))
	return ent
}
