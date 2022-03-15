package main

import (
	"github.com/pubgo/lava"
	"github.com/pubgo/lava/example/entry/gid"
	"github.com/pubgo/lava/example/entry/grpc_entry"
)

func main() {
	lava.Run(
		"service example",
		gid.GetEntry(),
		grpc_entry.GetEntry(),
	)
}
