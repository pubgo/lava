package main

import (
	"github.com/pubgo/lava"
	_ "github.com/pubgo/lava/clients/orm/driver/sqlite"
	"github.com/pubgo/lava/cmds/migratecmd"
	_ "github.com/pubgo/lava/core/registry/drivers/mdns"
	_ "github.com/pubgo/lava/debug/processhandler"
	"github.com/pubgo/lava/example/internal/bootstrap"
	"github.com/pubgo/lava/example/internal/cmds"
	"github.com/pubgo/lava/example/internal/migrates"
)

func main() {
	bootstrap.Init()
	lava.Run(cmds.Menu(), migratecmd.New(migrates.Migrations()))
}
