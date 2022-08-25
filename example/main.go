package main

import (
	"github.com/pubgo/funk/recovery"
	"github.com/pubgo/lava"
	_ "github.com/pubgo/lava/clients/orm/driver/sqlite"
	_ "github.com/pubgo/lava/core/registry/drivers/mdns"
	_ "github.com/pubgo/lava/debug/process"
	"github.com/pubgo/lava/example/internal/bootstrap"
	"github.com/pubgo/lava/example/internal/cmds"
)

func main() {
	defer recovery.Exit()

	var srv = lava.New()
	srv.Providers(bootstrap.Providers().Unwrap()...)
	lava.Run(srv, cmds.Menu())
}
