package main

import (
	"github.com/pubgo/lava"
	_ "github.com/pubgo/lava/clients/orm/driver/sqlite"
	_ "github.com/pubgo/lava/core/registry/drivers/mdns"
	_ "github.com/pubgo/lava/debug/process"

	"github.com/pubgo/lava/example/bootstrap"
	"github.com/pubgo/lava/example/handlers/gidhandler"
	"github.com/pubgo/lava/example/internal/cmds"
	"github.com/pubgo/lava/example/pkg/proto/gidpb"
)

func main() {
	var srv = lava.New()
	for _, p := range bootstrap.Providers() {
		srv.Provider(p)
	}

	gidpb.RegisterIdServer(srv, gidhandler.New())
	//hellopb.RegisterTestApiServer(srv, hellohandler.NewTestAPIHandler())
	//permpb.RegisterMenuServiceServer(srv, menurpc.New())
	//permpb.RegisterGroupServiceServer(srv, grouprpc.New())
	//permpb.RegisterOrgServiceServer(srv, orgrpc.New())
	//permpb.RegisterPermServiceServer(srv, permrpc.New())
	//permpb.RegisterRoleServiceServer(srv, rolerpc.New())

	lava.Run(srv, cmds.Menu())
}
