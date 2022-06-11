package main

import (
	"github.com/pubgo/lava"
	_ "github.com/pubgo/lava/clients/orm/driver/sqlite"
	_ "github.com/pubgo/lava/core/registry/drivers/mdns"
	"github.com/pubgo/lava/example/srv/gid"
	"github.com/pubgo/lava/example/srv/hello"
)

func main() {
	lava.Run(
		gid.NewSrv(),
		hello.NewSrv(),
	)
}
