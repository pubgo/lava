package main

import (
	"github.com/pubgo/lava"
	_ "github.com/pubgo/lava/clients/orm/driver/sqlite"
	_ "github.com/pubgo/lava/core/registry/drivers/mdns"
	"github.com/pubgo/lava/example/srv/gid"
	"github.com/pubgo/lava/example/srv/hello"
	"github.com/pubgo/lava/example/srv/perm"
	_ "github.com/pubgo/lava/logging/log_ext/klog"
	"github.com/pubgo/xerror"
)

func main() {
	defer xerror.RecoverAndExit()
	lava.Run(
		gid.NewSrv(),
		hello.NewSrv(),
		perm.NewSrv(),
	)
}
