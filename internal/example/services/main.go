package main

import (
	"github.com/pubgo/lava"
	"github.com/pubgo/lava/internal/example/services/entry/cli_entry"
	"github.com/pubgo/lava/internal/example/services/entry/gid"
	"github.com/pubgo/lava/internal/example/services/entry/gin_entry"
	"github.com/pubgo/lava/internal/example/services/entry/grpc_entry"
	"github.com/pubgo/lava/internal/example/services/entry/rest_entry"
	"github.com/pubgo/lava/internal/example/services/entry/task_entry"
)

func main() {
	lava.Run(
		"gid service",
		gid.GetEntry(),
		gin_entry.GetEntry(),
		cli_entry.GetEntry(),
		grpc_entry.GetEntry(),
		rest_entry.GetEntry(),
		task_entry.GetEntry(),
	)
}
