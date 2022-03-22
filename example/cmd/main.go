package main

import (
	"github.com/pubgo/lava"
	"github.com/pubgo/lava/example/entry/gid"
	"github.com/pubgo/lava/example/entry/grpc_entry"
)

import (
	_ "github.com/pubgo/lava/clients/orm/driver/sqlite"
	_ "github.com/pubgo/lava/plugins/metric/prometheus"
	_ "github.com/pubgo/lava/plugins/tracing/jaeger"
)

func main() {
	lava.Run(
		"service example",
		gid.GetEntry(),
		grpc_entry.GetEntry(),
	)
}
