package main

import (
	"github.com/pubgo/funk/v2/recovery"

	"github.com/pubgo/lava/v2/core/lavabuilder"
)

func main() {
	defer recovery.Exit()
	lavabuilder.Run(lavabuilder.New())
}
