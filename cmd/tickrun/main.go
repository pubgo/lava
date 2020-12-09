package main

import (
	"github.com/pubgo/golug"
	"github.com/pubgo/golug/cmd/tickrun/server"
	"github.com/pubgo/golug/cmd/tickrun/worker"
	"github.com/pubgo/xerror"
)

func main() {
	xerror.Exit(golug.Init())
	xerror.Exit(golug.Run(
		server.GetEntry(),
		worker.GetEntry(),
	))
}
