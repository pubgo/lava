package main

import (
	"github.com/pubgo/lava"
	_ "github.com/pubgo/lava/clients/orm/driver/sqlite"
	_ "github.com/pubgo/lava/core/metric/metric_driver/prometheus"
	"github.com/pubgo/lava/example/srv/gid"
	"github.com/pubgo/lava/example/srv/hello"
	_ "github.com/pubgo/lava/vars/vars_plugin"
	"go.uber.org/fx"
)

func main() {
	fx.Annotated{}
	lava.Run(
		gid.NewSrv(),
		hello.NewSrv(),
	)
}
