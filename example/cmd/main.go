package main

import (
	"github.com/pubgo/lava"
	"github.com/pubgo/lava/example/entry/gid"
	"github.com/pubgo/lava/example/entry/grpc_entry"

	_ "github.com/pubgo/lava/clients/orm/driver/sqlite"
	_ "github.com/pubgo/lava/core/metric/prometheus"
	_ "github.com/pubgo/lava/vars/vars_plugin"
)

func main() {
	lava.Run(
		"service example",
		gid.GetEntry(),
		grpc_entry.GetEntry(),
	)
}
