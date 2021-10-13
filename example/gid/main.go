package main

import (
	"github.com/pubgo/lava"
	"github.com/pubgo/lava/example/gid/entry/cli_entry"
	"github.com/pubgo/lava/example/gid/entry/gid"
	"github.com/pubgo/lava/example/gid/entry/grpc_entry"
	"github.com/pubgo/lava/example/gid/entry/rest_entry"
	"github.com/pubgo/lava/example/gid/entry/task_entry"
)

func main() {
	lava.Run(
		"gid service",
		gid.GetEntry(),
		cli_entry.GetEntry(),
		grpc_entry.GetEntry(),
		rest_entry.GetEntry(),
		task_entry.GetEntry(),
	)
}
