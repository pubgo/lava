package main

import (
	"github.com/pubgo/lava/cmds/runcmd"
	"github.com/pubgo/lava/internal/example/bootstrap"
)

func main() {
	bootstrap.Init()
	runcmd.Run()
}
