package main

import (
	"github.com/pubgo/lava"
	_ "github.com/pubgo/lava/clients/orm/driver/sqlite"
	_ "github.com/pubgo/lava/core/registry/drivers/mdns"
	_ "github.com/pubgo/lava/logging/log_ext/klog"

	"github.com/pubgo/lava/example/srv/gid"
	"github.com/pubgo/lava/example/srv/hello"
	"github.com/pubgo/lava/example/srv/perm"
)

func main() {
	lava.Run(
		gid.NewSrv(),
		hello.NewSrv(),
		perm.NewSrv(),
	)
}
