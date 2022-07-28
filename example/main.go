package main

import (
	"github.com/pubgo/funk/recovery"
	"github.com/pubgo/lava"
	_ "github.com/pubgo/lava/clients/orm/driver/sqlite"
	_ "github.com/pubgo/lava/core/registry/drivers/mdns"
	"github.com/pubgo/lava/example/handler"
	"github.com/pubgo/lava/example/pkg/proto/gidpb"
	_ "github.com/pubgo/lava/logging/log_ext/klog"
)

func main() {
	defer recovery.Exit()

	var srv = lava.NewSrv()

	gidpb.RegisterIdServer(srv, handler.NewId())

	lava.Run(srv)
}
