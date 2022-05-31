package main

import (
	"github.com/pubgo/lava"
	_ "github.com/pubgo/lava/clients/orm/driver/sqlite"
	"github.com/pubgo/lava/example/srv/gid"
	"github.com/pubgo/lava/example/srv/hello"
	"github.com/pubgo/lava/example/srv/web"
)

func main() {
	lava.Run(
		gid.NewSrv(),
		hello.NewSrv(),
		web.NewSrv(),
	)
}
