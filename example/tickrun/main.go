package main

import (
	"github.com/pubgo/golug"
	"github.com/pubgo/golug/example/tickrun/server"
	"github.com/pubgo/golug/example/tickrun/worker"
)

func main() {
	golug.Init()
	golug.Run(
		server.GetEntry(),
		worker.GetEntry(),
	)
}
