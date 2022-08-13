package main

import (
	"github.com/pubgo/funk/recovery"
	"github.com/pubgo/lava"
	_ "github.com/pubgo/lava/clients/orm/driver/sqlite"
	_ "github.com/pubgo/lava/core/registry/drivers/mdns"

	"github.com/pubgo/lava/example/handler"
	"github.com/pubgo/lava/example/internal/proto/gidpb"
)

func main() {
	defer recovery.Exit()

	var srv = lava.New()
	srv.RegisterGateway(gidpb.RegisterEchoServiceHandler)
	srv.RegisterServer(gidpb.RegisterIdServer, handler.NewId())
	lava.Run(srv)
}
