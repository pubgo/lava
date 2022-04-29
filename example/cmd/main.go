package main

import (
	"github.com/pubgo/lava"
	"github.com/pubgo/lava/example/srv/gid"
	"github.com/pubgo/lava/example/srv/hello"

	_ "github.com/pubgo/lava/clients/orm/driver/sqlite"
	_ "github.com/pubgo/lava/core/metric/metric_driver/prometheus"
	_ "github.com/pubgo/lava/module/registry_module"
)

func main() {
	lava.Run(
		gid.NewSrv(),
		hello.NewSrv(),
	)
}
