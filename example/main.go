package main

import (
	"github.com/pubgo/funk/recovery"
	"github.com/pubgo/lava"
	_ "github.com/pubgo/lava/clients/orm/driver/sqlite"
	_ "github.com/pubgo/lava/core/registry/drivers/mdns"
	_ "github.com/pubgo/lava/debug/process"
	"github.com/pubgo/lava/example/gen/proto/gidpb"
	"github.com/pubgo/lava/example/internal/bootstrap"
	"github.com/pubgo/lava/example/internal/handlers/gidrpc"

	"github.com/pubgo/lava/example/internal/cmds"
)

func main() {
	defer recovery.Exit()

	var srv = lava.New()
	srv.Providers(bootstrap.Providers().ToResult().Unwrap()...)

	gidpb.RegisterIdServer(srv, gidrpc.New())
	//hellopb.RegisterTestApiServer(srv, hellohandler.NewTestAPIHandler())
	//permpb.RegisterMenuServiceServer(srv, menurpc.New())
	//permpb.RegisterGroupServiceServer(srv, grouprpc.New())
	//permpb.RegisterOrgServiceServer(srv, orgrpc.New())
	//permpb.RegisterPermServiceServer(srv, permrpc.New())
	//permpb.RegisterRoleServiceServer(srv, rolerpc.New())

	lava.Run(srv, cmds.Menu())
}
