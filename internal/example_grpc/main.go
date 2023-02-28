package main

import (
	"github.com/pubgo/lava/cmds/running"
	"github.com/pubgo/lava/internal/example/bootstrap"
)

func main() {
	bootstrap.Init()
	running.Main()
}
