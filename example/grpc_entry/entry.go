package grpc_entry

import (
	"github.com/pubgo/golug"
	"github.com/pubgo/golug/example/grpc_entry/handler"
	"github.com/pubgo/golug/golug_entry"
)

func GetEntry() golug_entry.Entry {
	ent := golug.NewGrpcEntry("grpc", nil)
	ent.Version("v0.0.1")
	ent.Description("entry grpc test")
	ent.Register(handler.NewTestAPIHandler())
	return ent
}

func GetHttpEntry() golug_entry.Entry {
	ent := golug.NewRestEntry("grpc_api", nil)
	ent.Version("v0.0.1")
	ent.Description("entry grpc api test")
	ent.Register(handler.NewTestAPIHandler())
	return ent
}
